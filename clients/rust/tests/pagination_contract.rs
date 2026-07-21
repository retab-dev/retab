use std::collections::BTreeSet;
use std::fs;
use std::future::Future;
use std::io;
use std::sync::{Arc, Mutex};

use futures_util::{Stream, StreamExt};
use retab::{Error, Retab};
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::net::TcpListener;

// Cross-language pagination contract registry for the Rust SDK.
//
// A "list method" is `pub async fn list` OR any `pub async fn list_*`
// variant. An earlier version of the scan below matched `pub async fn list(`
// only, which silently excluded every `list_*` route — `usage.list_primitives`,
// `workflows.list_versions` and friends all return a cursor-paginated stream
// and carry the same auto-paging obligation.
//
// Labels follow the cross-language rule: the module filename alone for a bare
// `list`, and `module.rs::method` for a `list_*` variant.
const REGISTERED_LIST_METHODS: &[&str] = &[
    "classifications.rs",
    "edit_templates.rs",
    "edits.rs",
    "experiment_run_results.rs",
    "experiment_runs.rs",
    "extractions.rs",
    "files.rs",
    "parses.rs",
    "partitions.rs",
    "splits.rs",
    "usage.rs::list_blocks",
    "usage.rs::list_primitives",
    "usage.rs::list_runs",
    "workflow_artifacts.rs",
    "workflow_block_executions.rs",
    "workflow_blocks.rs",
    "workflow_blocks.rs::list_versions",
    "workflow_edges.rs",
    "workflow_edges.rs::list_versions",
    "workflow_experiments.rs",
    "workflow_review_versions.rs",
    "workflow_reviews.rs",
    "workflow_runs.rs",
    "workflow_steps.rs",
    "workflow_eval_run_results.rs",
    "workflow_eval_runs.rs",
    "workflow_evals.rs",
    "workflows.rs",
    "workflows.rs::list_versions",
];

// List methods that legitimately return a non-cursor envelope, so the
// auto-paging contract does not apply. Each entry carries its reason; stale
// entries fail `registry_covers_every_resource_list_method` alongside stale
// registry rows.
const NON_CURSOR_LIST_METHODS: &[(&str, &str)] = &[
    (
        "tables.rs",
        "list returns WorkflowTableListResponse ({tables: [...]}), no cursor envelope",
    ),
    (
        "secrets.rs::list_secrets",
        "returns SecretListResponse, an unpaginated envelope",
    ),
    (
        "secrets.rs::list_secret_value",
        "returns SecretValueResponse; 'list' is the API verb for a single value read, not a collection",
    ),
    (
        "workflows.rs::list_diff",
        "returns a single WorkflowGraphVersionDiff",
    ),
    (
        "workflow_blocks.rs::list_diff",
        "returns a single WorkflowBlockVersionDiff",
    ),
    (
        "workflow_edges.rs::list_diff",
        "returns a single WorkflowEdgeVersionDiff",
    ),
];

// The `list_*` variants that MUST stay registered. If the scan ever narrows
// back to `pub async fn list(`, these drop out silently and every remaining
// case still passes.
const PAGINATED_LIST_STAR_METHODS: &[&str] = &[
    "usage.rs::list_blocks",
    "usage.rs::list_primitives",
    "usage.rs::list_runs",
    "workflows.rs::list_versions",
    "workflow_blocks.rs::list_versions",
    "workflow_edges.rs::list_versions",
];

async fn start_two_page_server() -> (String, Arc<Mutex<Vec<String>>>) {
    let listener = TcpListener::bind("127.0.0.1:0").await.unwrap();
    let addr = listener.local_addr().unwrap();
    let calls = Arc::new(Mutex::new(Vec::new()));
    let captured = calls.clone();

    tokio::spawn(async move {
        for page in 0..2 {
            let Ok((mut socket, _)) = listener.accept().await else {
                return;
            };
            let mut buf = vec![0; 4096];
            let Ok(n) = socket.read(&mut buf).await else {
                return;
            };
            let request = String::from_utf8_lossy(&buf[..n]);
            let target = request
                .lines()
                .next()
                .and_then(|line| line.split_whitespace().nth(1))
                .unwrap_or("")
                .to_string();
            captured.lock().unwrap().push(target);

            let body = if page == 0 {
                r#"{"data":[],"list_metadata":{"before":null,"after":"cursor-2"}}"#
            } else {
                r#"{"data":[],"list_metadata":{"before":null,"after":null}}"#
            };
            let response = format!(
                "HTTP/1.1 200 OK\r\ncontent-type: application/json\r\ncontent-length: {}\r\nconnection: close\r\n\r\n{}",
                body.len(),
                body
            );
            let _ = socket.write_all(response.as_bytes()).await;
        }
    });

    (format!("http://{addr}"), calls)
}

async fn assert_auto_pages<T, P, Fut, Call>(service: &str, path: &str, call: Call)
where
    T: Unpin,
    P: Stream<Item = Result<T, Error>> + Unpin,
    Fut: Future<Output = Result<P, Error>>,
    Call: FnOnce(Retab) -> Fut,
{
    let (base_url, calls) = start_two_page_server().await;
    let client = Retab::builder()
        .api_key("test-api-key")
        .base_url(base_url)
        .build();

    let mut page = call(client).await.unwrap();
    while let Some(item) = page.next().await {
        item.unwrap();
    }

    let calls = calls.lock().unwrap().clone();
    assert_eq!(
        calls.len(),
        2,
        "{service}.list issued {} HTTP request(s); auto-pagination never followed `after`. This usually means list() bypassed Retab::request_page.",
        calls.len()
    );
    assert!(
        calls[0].starts_with(path),
        "{service}.list hit wrong path on first page: {:?}",
        calls
    );
    assert!(
        calls[1].starts_with(path) && calls[1].contains("after=cursor-2"),
        "{service}.list did not request the second page with after=cursor-2: {:?}",
        calls
    );
}

macro_rules! pagination_case {
    ($name:ident, $path:literal, |$client:ident| $body:expr) => {
        #[tokio::test]
        async fn $name() {
            assert_auto_pages(stringify!($name), $path, |$client| async move { $body }).await;
        }
    };
}

pagination_case!(
    classifications_list_auto_pages,
    "/v1/classifications",
    |client| {
        client
            .classifications()
            .list(retab::resources::classifications::ListParams::default())
            .await
    }
);
pagination_case!(
    edit_templates_list_auto_pages,
    "/v1/edits/templates",
    |client| {
        client
            .edits()
            .templates()
            .list(retab::resources::edit_templates::ListParams::default())
            .await
    }
);
pagination_case!(edits_list_auto_pages, "/v1/edits", |client| {
    client
        .edits()
        .list(retab::resources::edits::ListParams::default())
        .await
});
pagination_case!(
    experiment_run_results_list_auto_pages,
    "/v1/workflows/experiments/results",
    |client| {
        client
            .workflows()
            .experiments()
            .results()
            .list(retab::resources::experiment_run_results::ListParams::new(
                "run_x",
            ))
            .await
    }
);
pagination_case!(
    experiment_runs_list_auto_pages,
    "/v1/workflows/experiments/runs",
    |client| {
        client
            .workflows()
            .experiments()
            .runs()
            .list(retab::resources::experiment_runs::ListParams::default())
            .await
    }
);
pagination_case!(extractions_list_auto_pages, "/v1/extractions", |client| {
    client
        .extractions()
        .list(retab::resources::extractions::ListParams::default())
        .await
});
pagination_case!(files_list_auto_pages, "/v1/files", |client| {
    client
        .files()
        .list(retab::resources::files::ListParams::default())
        .await
});
pagination_case!(parses_list_auto_pages, "/v1/parses", |client| {
    client
        .parses()
        .list(retab::resources::parses::ListParams::default())
        .await
});
pagination_case!(partitions_list_auto_pages, "/v1/partitions", |client| {
    client
        .partitions()
        .list(retab::resources::partitions::ListParams::default())
        .await
});
pagination_case!(splits_list_auto_pages, "/v1/splits", |client| {
    client
        .splits()
        .list(retab::resources::splits::ListParams::default())
        .await
});
pagination_case!(
    workflow_artifacts_list_auto_pages,
    "/v1/workflows/artifacts",
    |client| {
        client
            .workflows()
            .artifacts()
            .list(retab::resources::workflow_artifacts::ListParams::default())
            .await
    }
);
pagination_case!(
    workflow_block_executions_list_auto_pages,
    "/v1/workflows/blocks/executions",
    |client| {
        client
            .workflows()
            .blocks()
            .executions()
            .list(retab::resources::workflow_block_executions::ListParams::new("run_x", "block_x"))
            .await
    }
);
pagination_case!(
    workflow_blocks_list_auto_pages,
    "/v1/workflows/blocks",
    |client| {
        client
            .workflows()
            .blocks()
            .list(retab::resources::workflow_blocks::ListParams::new("wf_x"))
            .await
    }
);
pagination_case!(
    workflow_edges_list_auto_pages,
    "/v1/workflows/edges",
    |client| {
        client
            .workflows()
            .edges()
            .list(retab::resources::workflow_edges::ListParams::new("wf_x"))
            .await
    }
);
pagination_case!(
    workflow_experiments_list_auto_pages,
    "/v1/workflows/experiments",
    |client| {
        client
            .workflows()
            .experiments()
            .list(retab::resources::workflow_experiments::ListParams::new(
                "wf_x",
            ))
            .await
    }
);
pagination_case!(
    workflow_review_versions_list_auto_pages,
    "/v1/workflows/reviews/versions",
    |client| {
        client
            .workflows()
            .reviews()
            .versions()
            .list(retab::resources::workflow_review_versions::ListParams::new(
                "rev_x",
            ))
            .await
    }
);
pagination_case!(
    workflow_reviews_list_auto_pages,
    "/v1/workflows/reviews",
    |client| {
        client
            .workflows()
            .reviews()
            .list(retab::resources::workflow_reviews::ListParams::default())
            .await
    }
);
pagination_case!(
    workflow_runs_list_auto_pages,
    "/v1/workflows/runs",
    |client| {
        client
            .workflows()
            .runs()
            .list(retab::resources::workflow_runs::ListParams::default())
            .await
    }
);
pagination_case!(
    workflow_steps_list_auto_pages,
    "/v1/workflows/steps",
    |client| {
        client
            .workflows()
            .steps()
            .list(retab::resources::workflow_steps::ListParams::default())
            .await
    }
);
pagination_case!(
    workflow_eval_run_results_list_auto_pages,
    "/v1/workflows/evals/results",
    |client| {
        client
            .workflows()
            .evals()
            .results()
            .list(retab::resources::workflow_eval_run_results::ListParams::new("run_x"))
            .await
    }
);
pagination_case!(
    workflow_eval_runs_list_auto_pages,
    "/v1/workflows/evals/runs",
    |client| {
        client
            .workflows()
            .evals()
            .runs()
            .list(retab::resources::workflow_eval_runs::ListParams::default())
            .await
    }
);
pagination_case!(
    workflow_evals_list_auto_pages,
    "/v1/workflows/evals",
    |client| {
        client
            .workflows()
            .evals()
            .list(retab::resources::workflow_evals::ListParams::new("wf_x"))
            .await
    }
);
pagination_case!(workflows_list_auto_pages, "/v1/workflows", |client| {
    client
        .workflows()
        .list(retab::resources::workflows::ListParams::default())
        .await
});
pagination_case!(
    workflows_list_versions_auto_pages,
    "/v1/workflows/versions",
    |client| {
        client
            .workflows()
            .list_versions(retab::resources::workflows::ListVersionsParams::new("wf_x"))
            .await
    }
);
pagination_case!(
    workflow_blocks_list_versions_auto_pages,
    "/v1/workflows/blocks/versions",
    |client| {
        client
            .workflows()
            .blocks()
            .list_versions(retab::resources::workflow_blocks::ListVersionsParams::new(
                "wf_x",
            ))
            .await
    }
);
pagination_case!(
    workflow_edges_list_versions_auto_pages,
    "/v1/workflows/edges/versions",
    |client| {
        client
            .workflows()
            .edges()
            .list_versions(retab::resources::workflow_edges::ListVersionsParams::new(
                "wf_x",
            ))
            .await
    }
);
pagination_case!(usage_list_blocks_auto_pages, "/v1/usage/blocks", |client| {
    client
        .usage()
        .list_blocks(retab::resources::usage::ListBlocksParams::default())
        .await
});
pagination_case!(
    usage_list_primitives_auto_pages,
    "/v1/usage/primitives",
    |client| {
        client
            .usage()
            .list_primitives(retab::resources::usage::ListPrimitivesParams::default())
            .await
    }
);
pagination_case!(usage_list_runs_auto_pages, "/v1/usage/runs", |client| {
    client
        .usage()
        .list_runs(retab::resources::usage::ListRunsParams::default())
        .await
});

/// Scan `src/resources/` and return the discovery label of every list method:
/// the module filename for a bare `pub async fn list`, and `module.rs::method`
/// for every `pub async fn list_*` variant.
///
/// The `*_with_options` siblings are the same route with an explicit
/// `RequestOptions` argument (every list method has one), not routes of their
/// own, so they are excluded by name.
fn discover_list_methods() -> io::Result<BTreeSet<String>> {
    let resources_dir = std::path::Path::new(env!("CARGO_MANIFEST_DIR"))
        .join("src")
        .join("resources");
    let mut discovered = BTreeSet::new();

    for entry in fs::read_dir(resources_dir)? {
        let entry = entry?;
        let path = entry.path();
        if path.extension().and_then(|ext| ext.to_str()) != Some("rs") {
            continue;
        }
        let Some(file_name) = path.file_name().and_then(|name| name.to_str()) else {
            continue;
        };
        let contents = fs::read_to_string(&path)?;
        for line in contents.lines() {
            const PREFIX: &str = "pub async fn ";
            let Some(rest) = line.trim_start().strip_prefix(PREFIX) else {
                continue;
            };
            let Some(name) = rest.split('(').next() else {
                continue;
            };
            if name != "list" && !name.starts_with("list_") {
                continue;
            }
            if name.ends_with("_with_options") {
                continue;
            }
            if name == "list" {
                discovered.insert(file_name.to_string());
            } else {
                discovered.insert(format!("{file_name}::{name}"));
            }
        }
    }

    Ok(discovered)
}

#[test]
fn registry_covers_every_resource_list_method() -> io::Result<()> {
    let registered: BTreeSet<&str> = REGISTERED_LIST_METHODS.iter().copied().collect();
    let non_cursor: BTreeSet<&str> = NON_CURSOR_LIST_METHODS
        .iter()
        .map(|(label, _reason)| *label)
        .collect();
    let discovered = discover_list_methods()?;

    let missing: Vec<_> = discovered
        .iter()
        .filter(|name| !registered.contains(name.as_str()) && !non_cursor.contains(name.as_str()))
        .cloned()
        .collect();
    let stale: Vec<_> = registered
        .iter()
        .chain(non_cursor.iter())
        .filter(|name| !discovered.contains(**name))
        .copied()
        .collect();

    assert!(
        missing.is_empty(),
        "List methods missing from the pagination contract registry: {missing:?}. \
         Register them, or add them to NON_CURSOR_LIST_METHODS with a reason."
    );
    assert!(
        stale.is_empty(),
        "Pagination contract entries without a matching list method: {stale:?}"
    );
    Ok(())
}

/// Pin the widened scan. The bug this guards is a discovery walk that narrows
/// back to `pub async fn list(`: every `list_*` route would drop out of the
/// registry check and the suite would still be green.
#[test]
fn discovery_covers_list_star_variants() -> io::Result<()> {
    let discovered = discover_list_methods()?;
    let missing: Vec<_> = PAGINATED_LIST_STAR_METHODS
        .iter()
        .filter(|name| !discovered.contains(**name))
        .copied()
        .collect();
    assert!(
        missing.is_empty(),
        "These list_* methods were not discovered: {missing:?}. \
         The scan regressed to matching only `pub async fn list(`."
    );
    Ok(())
}

/// Every non-cursor exemption needs a reason, so the allowlist documents
/// itself rather than accumulating bare strings.
#[test]
fn non_cursor_entries_are_documented() {
    for (label, reason) in NON_CURSOR_LIST_METHODS {
        assert!(
            !reason.trim().is_empty(),
            "NON_CURSOR_LIST_METHODS entry {label:?} needs a reason explaining why the cursor contract does not apply"
        );
    }
}
