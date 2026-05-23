use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};

use futures_util::StreamExt;
use retab::enums::{UpdateWorkflowBlockRequestConfigMode, WorkflowBlockCreateRequestType};
use retab::models::{
    CancelWorkflowRequest, CreateFreshWorkflowRunRequest, CreateWorkflowRequest,
    CreateWorkflowRunRequest, DeclarativeWorkflowRequest, PublishWorkflowRequest,
    UpdateWorkflowBlockRequest, UpdateWorkflowRequest, WorkflowBlockCreateRequest,
    WorkflowEdgeCreateRequest, WorkflowGraphDiagnosisRequest,
};
use retab::resources;
use retab::Retab;
use serde_json::json;

#[tokio::test]
#[ignore = "requires local Retab backend and RETAB_API_KEY"]
async fn exercise_workflow_system_live() {
    macro_rules! api {
        ($expr:expr, $context:literal) => {
            $expr
                .await
                .map_err(|err| format!("{}: {err:?}", $context))?
        };
    }

    let api_key = std::env::var("RETAB_API_KEY").expect("RETAB_API_KEY must be set");
    let base_url =
        std::env::var("RETAB_BASE_URL").unwrap_or_else(|_| "http://localhost:4000".to_string());
    let client = Retab::builder().api_key(api_key).base_url(base_url).build();
    let unique = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap()
        .as_millis();

    let workflow_body = CreateWorkflowRequest {
        name: Some(format!("rust-sdk-live-probe-{unique}")),
        description: Some("temporary workflow created by Rust SDK live probe".to_string()),
    };

    let workflow = client
        .workflows()
        .create(resources::workflows::CreateParams::new(workflow_body))
        .await
        .expect("create workflow");
    println!("created workflow {}", workflow.id);

    let probe_result: Result<(), String> = (async {
        let update_body = UpdateWorkflowRequest {
            description: Some("updated by Rust SDK live probe".to_string()),
            ..Default::default()
        };
        let updated = api!(
            client.workflows().update(
                &workflow.id,
                resources::workflows::UpdateParams::new(update_body),
            ),
            "update workflow"
        );
        if updated.description.as_deref() != Some("updated by Rust SDK live probe") {
            return Err("update workflow returned unexpected description".to_string());
        }

        let fetched = api!(client.workflows().get(&workflow.id), "get workflow");
        if fetched.id != workflow.id {
            return Err("get workflow returned wrong id".to_string());
        }

        let mut workflows_page = api!(
            client.workflows().list(resources::workflows::ListParams {
                limit: Some(1),
                ..Default::default()
            }),
            "list workflows"
        );
        let _ = workflows_page
            .next()
            .await
            .transpose()
            .map_err(|err| format!("stream workflows: {err:?}"))?;

        let initial_blocks = api!(
            client
                .workflow_blocks()
                .list(resources::workflow_blocks::ListParams::new(&workflow.id)),
            "list initial blocks"
        );
        println!("initial block count {}", initial_blocks.data.len());

        let mut note_body =
            WorkflowBlockCreateRequest::new(&workflow.id, WorkflowBlockCreateRequestType::Note);
        note_body.label = Some("Rust SDK probe note".to_string());
        note_body.position_x = Some(320.0);
        note_body.position_y = Some(120.0);
        note_body.config = Some(HashMap::from([(
            "text".to_string(),
            json!("Created by Rust SDK live probe"),
        )]));
        let note = api!(
            client
                .workflow_blocks()
                .create(resources::workflow_blocks::CreateParams::new(note_body)),
            "create note block"
        );
        println!("created note block {}", note.id);

        let block_patch = UpdateWorkflowBlockRequest {
            label: Some("Rust SDK probe note updated".to_string()),
            position_x: Some(360.0),
            config_mode: Some(UpdateWorkflowBlockRequestConfigMode::Merge),
            config: Some(HashMap::from([(
                "text".to_string(),
                json!("Updated by Rust SDK live probe"),
            )])),
            ..Default::default()
        };
        let updated_note = api!(
            client.workflow_blocks().update(
                &note.id,
                resources::workflow_blocks::UpdateParams {
                    workflow_id: Some(workflow.id.clone()),
                    body: block_patch,
                },
            ),
            "update note block"
        );
        if updated_note.label.as_deref() != Some("Rust SDK probe note updated") {
            return Err("update note block returned unexpected label".to_string());
        }

        let _ = api!(
            client.workflow_blocks().get(
                &note.id,
                resources::workflow_blocks::GetParams {
                    workflow_id: Some(workflow.id.clone()),
                },
            ),
            "get note block"
        );

        let blocks_after_note = api!(
            client
                .workflow_blocks()
                .list(resources::workflow_blocks::ListParams::new(&workflow.id)),
            "list blocks after note"
        );
        println!("block count after note {}", blocks_after_note.data.len());

        if let Some(source) = blocks_after_note.data.iter().find(|b| b.id != note.id) {
            let edge_body = WorkflowEdgeCreateRequest::new(&workflow.id, &source.id, &note.id);
            match client
                .workflow_edges()
                .create(resources::workflow_edges::CreateParams::new(edge_body))
                .await
            {
                Ok(edge) => {
                    println!("created edge {}", edge.id);
                    let got = api!(client.workflow_edges().get(&edge.id), "get edge");
                    if got.id != edge.id {
                        return Err("get edge returned wrong id".to_string());
                    }
                    api!(client.workflow_edges().delete(&edge.id), "delete edge");
                }
                Err(err) => {
                    println!("edge create returned API error, continuing: {err:?}");
                }
            }
        }

        let _diagnosis = api!(
            client.workflows().diagnose(
                &workflow.id,
                resources::workflows::DiagnoseParams::new(WorkflowGraphDiagnosisRequest::default()),
            ),
            "diagnose persisted workflow"
        );

        let exported_spec = api!(
            client.workflow_specs().get(&workflow.id),
            "export workflow spec"
        );
        if exported_spec.workflow_id != workflow.id {
            return Err("export workflow spec returned wrong workflow_id".to_string());
        }
        println!(
            "exported spec bytes {}",
            exported_spec.yaml_definition.len()
        );

        let spec_body = DeclarativeWorkflowRequest::new(exported_spec.yaml_definition.clone());
        match client
            .workflow_specs()
            .validate(resources::workflow_specs::ValidateParams::new(
                spec_body.clone(),
            ))
            .await
        {
            Ok(validation) => {
                println!(
                    "validated spec blocks={} edges={} valid={}",
                    validation.block_count, validation.edge_count, validation.is_valid
                );

                let plan = api!(
                    client
                        .workflow_specs()
                        .plan(resources::workflow_specs::PlanParams::new(
                            spec_body.clone()
                        )),
                    "plan exported workflow spec"
                );
                println!("planned exported spec action={:?}", plan.action);

                let apply = api!(
                    client
                        .workflow_specs()
                        .apply(resources::workflow_specs::ApplyParams::new(spec_body)),
                    "apply exported workflow spec"
                );
                println!("applied exported spec action={:?}", apply.action);
            }
            Err(err) => println!("validate exported spec returned API error, continuing: {err:?}"),
        }

        match client
            .workflows()
            .publish(
                &workflow.id,
                resources::workflows::PublishParams {
                    body: Some(PublishWorkflowRequest {
                        description: Some("Rust SDK live probe publication".to_string()),
                    }),
                },
            )
            .await
        {
            Ok(published) => println!("published workflow {:?}", published.published),
            Err(err) => println!("publish returned API error, continuing: {err:?}"),
        }

        let mut fresh_run = CreateFreshWorkflowRunRequest::new(&workflow.id);
        fresh_run.version = Some("draft".to_string());
        match client
            .workflow_runs()
            .create(resources::workflow_runs::CreateParams::new(
                CreateWorkflowRunRequest::from(fresh_run),
            ))
            .await
        {
            Ok(run) => {
                println!("created run {}", run.id);
                let _ = api!(client.workflow_runs().get(&run.id), "get workflow run");
                let cancel_body = CancelWorkflowRequest {
                    command_id: Some(format!("rust-sdk-live-probe-{unique}")),
                };
                let _ = client
                    .workflow_runs()
                    .cancel(
                        &run.id,
                        resources::workflow_runs::CancelParams {
                            body: Some(cancel_body),
                        },
                    )
                    .await;
            }
            Err(err) => println!("run create returned API error, continuing: {err:?}"),
        }

        api!(
            client.workflow_blocks().delete(
                &note.id,
                resources::workflow_blocks::DeleteParams {
                    workflow_id: Some(workflow.id.clone()),
                },
            ),
            "delete note block"
        );

        Ok(())
    })
    .await;

    let delete_result = client.workflows().delete(&workflow.id).await;
    if let Err(err) = delete_result {
        panic!("delete workflow cleanup failed: {err:?}");
    }
    if let Err(err) = probe_result {
        panic!("{err}");
    }
}
