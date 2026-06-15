use std::collections::HashMap;
use std::time::{SystemTime, UNIX_EPOCH};

use futures_util::StreamExt;
use retab::enums::{
    DeclarativeApplyResponseAction, DeclarativePlanResponseAction, WorkflowBlockType,
};
use retab::models::{
    CreateWorkflowRunRequest, DeclarativeWorkflowRequest, PublishWorkflowRequest,
    UpdateWorkflowRequest,
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
    let workflow_id = format!("wrk_rust_sdk_live_probe_{unique}");

    let yaml_definition = format!(
        r#"apiVersion: workflows.retab.com/v1alpha2
kind: Workflow
metadata:
  id: {workflow_id}
  name: Rust SDK Invoice Validation Probe {unique}
  description: Temporary invoice total validation workflow created by the Rust SDK live probe.
spec:
  blocks:
    start:
      type: start_json
      label: Invoice JSON
      config:
        json_schema:
          type: object
          properties:
            invoice_id:
              type: string
            line_items:
              type: array
              items:
                type: object
                properties:
                  description:
                    type: string
                  amount:
                    type: number
                required:
                  - description
                  - amount
            tax_rate:
              type: number
            stated_total:
              type: number
          required:
            - invoice_id
            - line_items
            - tax_rate
            - stated_total
    transform:
      type: function
      label: Validate Invoice Total
      config:
        output_schema:
          type: object
          properties:
            invoice_id:
              type: string
            subtotal:
              type: number
            tax:
              type: number
            computed_total:
              type: number
            stated_total:
              type: number
            is_valid:
              type: boolean
            error_message:
              type: string
          required:
            - invoice_id
            - subtotal
            - tax
            - computed_total
            - stated_total
            - is_valid
            - error_message
        code: |
          from models import Input, Output

          def transform(input_data: Input) -> Output:
              subtotal = sum(item.amount for item in input_data.line_items)
              tax = round(subtotal * input_data.tax_rate, 2)
              computed_total = round(subtotal + tax, 2)
              stated_total = round(input_data.stated_total, 2)
              is_valid = abs(computed_total - stated_total) <= 0.01
              error_message = "" if is_valid else f"Expected {{computed_total}}, got {{stated_total}}"
              return Output(
                  invoice_id=input_data.invoice_id,
                  subtotal=subtotal,
                  tax=tax,
                  computed_total=computed_total,
                  stated_total=stated_total,
                  is_valid=is_valid,
                  error_message=error_message,
              )
  edges:
    - from:
        block: start
        handle: output-json-0
      to:
        block: transform
        handle: input-json-0
"#
    );

    let probe_result: Result<(), String> = (async {
        let spec_body = DeclarativeWorkflowRequest::new(yaml_definition.clone());

        let validation = api!(
            client
                .workflows()
                .spec()
                .validate(resources::workflow_spec::ValidateParams::new(
                    spec_body.clone(),
                )),
            "validate workflow spec"
        );
        if !validation.is_valid {
            return Err(format!(
                "valid workflow spec returned is_valid=false: {:?}",
                validation.diagnostics
            ));
        }
        if validation.block_count != 2 || validation.edge_count != 1 {
            return Err(format!(
                "valid workflow spec returned unexpected topology counts: blocks={}, edges={}",
                validation.block_count, validation.edge_count
            ));
        }

        let plan = api!(
            client
                .workflows()
                .plan(resources::workflows::PlanParams::new(spec_body.clone())),
            "plan workflow spec"
        );
        if plan.action != DeclarativePlanResponseAction::Create {
            return Err(format!(
                "new workflow spec planned unexpected action: {:?}",
                plan.action
            ));
        }

        let apply = api!(
            client
                .workflows()
                .apply(resources::workflows::ApplyParams::new(spec_body)),
            "apply workflow spec"
        );
        if apply.workflow_id != workflow_id {
            return Err("apply workflow spec returned wrong workflow_id".to_string());
        }
        if apply.action != DeclarativeApplyResponseAction::Create || !apply.created {
            return Err(format!(
                "apply workflow spec returned unexpected action/created: {:?}/{}",
                apply.action, apply.created
            ));
        }
        println!("created workflow {workflow_id} from declarative spec");

        let update_body = UpdateWorkflowRequest {
            description: Some("validates invoice subtotals, tax, and stated totals".to_string()),
            ..Default::default()
        };
        let updated = api!(
            client.workflows().update(
                &workflow_id,
                resources::workflows::UpdateParams::new(update_body),
            ),
            "update workflow"
        );
        if updated.description.as_deref()
            != Some("validates invoice subtotals, tax, and stated totals")
        {
            return Err("update workflow returned unexpected description".to_string());
        }

        let fetched = api!(client.workflows().get(&workflow_id), "get workflow");
        if fetched.id != workflow_id {
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

        let blocks = api!(
            client
                .workflows()
                .blocks()
                .list(resources::workflow_blocks::ListParams::new(&workflow_id)),
            "list workflow blocks"
        );
        let start_json_id = blocks
            .data
            .iter()
            .find(|block| block.type_ == WorkflowBlockType::StartJson)
            .map(|block| block.id.clone())
            .ok_or_else(|| "applied workflow has no start_json block".to_string())?;
        let transform_id = blocks
            .data
            .iter()
            .find(|block| block.type_ == WorkflowBlockType::Function)
            .map(|block| block.id.clone())
            .ok_or_else(|| "applied workflow has no function block".to_string())?;
        let _ = api!(
            client.workflows().blocks().get(
                &transform_id,
                resources::workflow_blocks::GetParams {
                    workflow_id: Some(workflow_id.clone()),
                },
            ),
            "get function block"
        );

        let edges = api!(
            client
                .workflows()
                .edges()
                .list(resources::workflow_edges::ListParams::new(&workflow_id)),
            "list workflow edges"
        );
        if edges.data.len() != 1 {
            return Err(format!(
                "expected one workflow edge, got {}",
                edges.data.len()
            ));
        }
        let edge = api!(
            client.workflows().edges().get(&edges.data[0].id),
            "get workflow edge"
        );
        if edge.workflow_id != workflow_id {
            return Err("get workflow edge returned wrong workflow_id".to_string());
        }

        let exported_spec = api!(
            client.workflows().spec().get(&workflow_id),
            "export workflow spec"
        );
        if exported_spec.workflow_id != workflow_id {
            return Err("export workflow spec returned wrong workflow_id".to_string());
        }
        let exported_body = DeclarativeWorkflowRequest::new(exported_spec.yaml_definition);
        let exported_plan = api!(
            client
                .workflows()
                .plan(resources::workflows::PlanParams::new(exported_body.clone(),)),
            "plan exported workflow spec"
        );
        if exported_plan.action != DeclarativePlanResponseAction::Noop {
            return Err(format!(
                "exported workflow spec planned unexpected action: {:?}",
                exported_plan.action
            ));
        }
        let exported_apply = api!(
            client
                .workflows()
                .apply(resources::workflows::ApplyParams::new(exported_body)),
            "apply exported workflow spec"
        );
        if exported_apply.action != DeclarativeApplyResponseAction::Noop {
            return Err(format!(
                "exported workflow spec applied unexpected action: {:?}",
                exported_apply.action
            ));
        }

        let published = api!(
            client.workflows().publish(
                &workflow_id,
                resources::workflows::PublishParams {
                    body: Some(PublishWorkflowRequest {
                        description: Some("Rust SDK live probe publication".to_string()),
                    }),
                },
            ),
            "publish workflow"
        );
        let published_version_id = published
            .published
            .and_then(|published| published.version_id)
            .ok_or_else(|| "publish workflow returned no version id".to_string())?;
        println!("published workflow {workflow_id} as {published_version_id}");

        let mut fresh_run = CreateWorkflowRunRequest::new(&workflow_id);
        fresh_run.version = Some("production".to_string());
        fresh_run.json_inputs = Some(HashMap::from([(
            start_json_id,
            json!({
                "invoice_id": format!("inv-rust-sdk-live-probe-{unique}"),
                "line_items": [
                    {"description": "warehouse handling", "amount": 120.0},
                    {"description": "local delivery", "amount": 80.0}
                ],
                "tax_rate": 0.2,
                "stated_total": 240.0
            }),
        )]));
        let run = api!(
            client
                .workflows()
                .runs()
                .create(resources::workflow_runs::CreateParams::new(fresh_run)),
            "create workflow run"
        );
        println!("created workflow run {}", run.id);
        let got_run = api!(client.workflows().runs().get(&run.id), "get workflow run");
        if got_run.id != run.id {
            return Err("get workflow run returned wrong id".to_string());
        }

        let mut runs_page = api!(
            client
                .workflows()
                .runs()
                .list(resources::workflow_runs::ListParams {
                    workflow_id: Some(workflow_id.clone()),
                    limit: Some(1),
                    ..Default::default()
                }),
            "list workflow runs"
        );
        let _ = runs_page
            .next()
            .await
            .transpose()
            .map_err(|err| format!("stream workflow runs: {err:?}"))?;

        api!(
            client.workflows().runs().delete(&run.id),
            "delete workflow run"
        );

        Ok(())
    })
    .await;

    let delete_result = client.workflows().delete(&workflow_id).await;
    if let Err(err) = delete_result {
        panic!("delete workflow cleanup failed: {err:?}");
    }
    if let Err(err) = probe_result {
        panic!("{err}");
    }
}
