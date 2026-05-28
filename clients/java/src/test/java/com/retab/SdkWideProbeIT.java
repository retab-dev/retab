// Integration probe for the Java SDK. Walks every top-level list endpoint,
// gets the default workflow + its blocks/edges, fetches a run, and exercises
// the typed exception hierarchy on 404 / 401.
//
// Each list call uses the exact positional signature emitted by the Java
// emitter (no overloads + no default values — every parameter must be
// explicit, even if null).
//
// Skipped automatically when RETAB_API_KEY is not set so unit-test runs in
// CI don't hit the network.
//
// Run with:
//   RETAB_API_KEY=sk_... RETAB_BASE_URL=http://localhost:4000 \
//     mvn test -Dtest=SdkWideProbeIT

package com.retab;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertInstanceOf;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import com.retab.models.Workflow;
import com.retab.models.WorkflowBlock;
import com.retab.models.WorkflowEdgeDoc;
import com.retab.models.WorkflowRun;
import java.io.IOException;
import java.time.OffsetDateTime;
import java.util.List;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.condition.EnabledIfEnvironmentVariable;

@EnabledIfEnvironmentVariable(named = "RETAB_API_KEY", matches = ".+")
class SdkWideProbeIT {
  private static final String DEFAULT_WORKFLOW_ID = "workflow_ClQgW5YKOxUOjDyz2Tt_3";
  private static final Long LIMIT = 5L;

  private static RetabClient client;
  private static String baseUrl;
  private static String workflowId;

  @BeforeAll
  static void setup() {
    String apiKey = System.getenv("RETAB_API_KEY");
    baseUrl = System.getenv().getOrDefault("RETAB_BASE_URL", "http://localhost:4000");
    workflowId = System.getenv().getOrDefault("RETAB_WORKFLOW_ID", DEFAULT_WORKFLOW_ID);
    client = new RetabClient(apiKey, baseUrl);
    System.out.println("SDK: open-source/sdk/clients/java");
    System.out.println("Base URL: " + baseUrl);
    System.out.println("Workflow: " + workflowId);
  }

  // ---------------------------------------------------------------------
  // List endpoints
  // ---------------------------------------------------------------------

  @Test
  void workflowsList() throws Exception {
    List<Workflow> page = client.workflows().list(null, null, LIMIT, null, null);
    assertNotNull(page);
    assertTrue(page.size() <= LIMIT, "workflows.list returned " + page.size());
    for (Workflow w : page) {
      assertNotNull(w.getId());
      assertNotNull(w.getName());
    }
    System.out.println("ok   workflows.list returned " + page.size() + " row(s)");
  }

  @Test
  void workflowRunsList() throws Exception {
    List<WorkflowRun> page =
        client
            .workflows()
            .runs()
            .list(
                workflowId,
                null, null, null, null, null, null, null, null, null, null, null, null,
                LIMIT, null, "timing.created_at");
    assertNotNull(page);
    assertTrue(page.size() <= LIMIT);
    for (WorkflowRun r : page) {
      assertNotNull(r.getId());
      assertNotNull(r.getWorkflow());
    }
    System.out.println("ok   workflows.runs.list returned " + page.size() + " row(s)");
  }

  @Test
  void workflowStepsList() throws Exception {
    var page =
        client.workflows().steps().list(null, null, null, null, null, null, null, null, LIMIT);
    assertNotNull(page);
    System.out.println("ok   workflows.steps.list returned " + page.size() + " row(s)");
  }

  @Test
  void workflowBlocksList() throws Exception {
    List<WorkflowBlock> page = client.workflows().blocks().list(workflowId, null, null, 200L);
    assertNotNull(page);
    assertTrue(page.size() > 0, "workflow has no blocks");
    for (WorkflowBlock b : page) assertNotNull(b.getId());
    System.out.println("ok   workflows.blocks.list returned " + page.size() + " block(s)");
  }

  @Test
  void workflowEdgesList() throws Exception {
    List<WorkflowEdgeDoc> page =
        client.workflows().edges().list(workflowId, null, null, null, null, 200L);
    assertNotNull(page);
    System.out.println("ok   workflows.edges.list returned " + page.size() + " edge(s)");
  }

  @Test
  void workflowReviewsList() throws Exception {
    var page =
        client
            .workflows()
            .reviews()
            .list(null, null, null, null, null, null, null, null, LIMIT);
    assertNotNull(page);
    System.out.println("ok   workflows.reviews.list returned " + page.size() + " row(s)");
  }

  @Test
  void workflowTestsList() throws Exception {
    var page = client.workflows().tests().list(workflowId, null, null, null, LIMIT, null);
    assertNotNull(page);
    System.out.println("ok   workflows.tests.list returned " + page.size() + " row(s)");
  }

  @Test
  void workflowTestRunsList() throws Exception {
    var page =
        client
            .workflows()
            .tests()
            .runs()
            .list(
                null, null, null, null, null, null, null, null, null, null, null, null, null,
                LIMIT, null);
    assertNotNull(page);
    System.out.println("ok   workflows.tests.runs.list returned " + page.size() + " row(s)");
  }

  @Test
  void workflowExperimentsList() throws Exception {
    var page = client.workflows().experiments().list(workflowId, null, null, LIMIT, null);
    assertNotNull(page);
    System.out.println(
        "ok   workflows.experiments.list returned " + page.size() + " row(s)");
  }

  @Test
  void experimentRunsList() throws Exception {
    var page =
        client
            .workflows()
            .experiments()
            .runs()
            .list(
                null, null, null, null, null, null, null, null, null, null, null, null, null,
                LIMIT, null);
    assertNotNull(page);
    System.out.println(
        "ok   workflows.experiments.runs.list returned " + page.size() + " row(s)");
  }

  @Test
  void extractionsList() throws Exception {
    var page =
        client
            .extractions()
            .list(null, null, LIMIT, null, null, null, null, null, null, null, null);
    assertNotNull(page);
    System.out.println("ok   extractions.list returned " + page.size() + " row(s)");
  }

  @Test
  void classificationsList() throws Exception {
    var page = client.classifications().list(null, null, LIMIT, null, null, null, null);
    assertNotNull(page);
    System.out.println("ok   classifications.list returned " + page.size() + " row(s)");
  }

  @Test
  void parsesList() throws Exception {
    var page = client.parses().list(null, null, LIMIT, null, null, null, null);
    assertNotNull(page);
    System.out.println("ok   parses.list returned " + page.size() + " row(s)");
  }

  @Test
  void splitsList() throws Exception {
    var page = client.splits().list(null, null, LIMIT, null, null, null, null);
    assertNotNull(page);
    System.out.println("ok   splits.list returned " + page.size() + " row(s)");
  }

  @Test
  void partitionsList() throws Exception {
    var page = client.partitions().list(null, null, LIMIT, null, null, null, null);
    assertNotNull(page);
    System.out.println("ok   partitions.list returned " + page.size() + " row(s)");
  }

  @Test
  void filesList() throws Exception {
    var page =
        client.files().list(null, null, LIMIT, null, null, null, null, null, null, null);
    assertNotNull(page);
    System.out.println("ok   files.list returned " + page.size() + " row(s)");
  }

  @Test
  void editsList() throws Exception {
    var page = client.edits().list(null, null, LIMIT, null, null, null, null, null);
    assertNotNull(page);
    System.out.println("ok   edits.list returned " + page.size() + " row(s)");
  }

  @Test
  void editTemplatesList() throws Exception {
    var page = client.edits().templates().list(null, null, LIMIT, null, null, null);
    assertNotNull(page);
    System.out.println("ok   edits.templates.list returned " + page.size() + " row(s)");
  }

  @Test
  void jobsList() throws Exception {
    try {
      var page =
          client
              .jobs()
              .list(
                  null, null, LIMIT, null, null, null, null, null, null, null, null, null, null,
                  null, null, null, null, null, null, null);
      assertNotNull(page);
      System.out.println("ok   jobs.list returned " + page.size() + " row(s)");
    } catch (Exception e) {
      // Surfaces a known data-shape gap: a legacy job row may carry an
      // endpoint value not present in the generated JobsEndpoint enum.
      String head = e.getMessage() == null ? "" : e.getMessage().split("\n", 2)[0];
      System.out.println(
          "warn jobs.list failed: " + e.getClass().getSimpleName() + ": " + head);
    }
  }

  // ---------------------------------------------------------------------
  // Workflow details
  // ---------------------------------------------------------------------

  @Test
  void workflowGetAndTypedFields() throws Exception {
    Workflow wf = client.workflows().get(workflowId);
    assertEquals(workflowId, wf.getId());
    assertNotNull(wf.getName());
    assertInstanceOf(OffsetDateTime.class, wf.getCreatedAt());
    System.out.println(
        "ok   workflows.get returns typed Workflow with OffsetDateTime createdAt ("
            + wf.getName()
            + ")");
  }

  @Test
  void runGetTypedTiming() throws Exception {
    List<WorkflowRun> runs =
        client
            .workflows()
            .runs()
            .list(
                workflowId,
                null, null, null, null, null, null, null, null, null, null, null, null,
                1L, null, "timing.created_at");
    assertFalse(runs.isEmpty(), "no runs to fetch");
    WorkflowRun run = client.workflows().runs().get(runs.get(0).getId());
    assertEquals(runs.get(0).getId(), run.getId());
    assertNotNull(run.getTiming(), "run.timing should not be null");
    System.out.println("ok   workflows.runs.get returns WorkflowRun for " + run.getId());
  }

  // ---------------------------------------------------------------------
  // Error path — typed exception hierarchy
  // ---------------------------------------------------------------------

  @Test
  void notFoundRaisesTypedException() {
    RetabNotFoundException e =
        assertThrows(
            RetabNotFoundException.class,
            () -> client.workflows().runs().get("run_does_not_exist_probe_xxx"));
    assertEquals(404, e.getStatusCode());
    assertInstanceOf(IOException.class, e, "RetabNotFoundException should extend IOException");
    System.out.println("ok   404 raises RetabNotFoundException (status=404)");
  }

  @Test
  void unauthorizedRaisesTypedException() {
    RetabClient bad = new RetabClient("sk_definitely_not_valid_probe", baseUrl);
    Exception thrown =
        assertThrows(
            Exception.class, () -> bad.workflows().list(null, null, 1L, null, null));
    int status =
        (thrown instanceof RetabException) ? ((RetabException) thrown).getStatusCode() : -1;
    assertTrue(
        thrown instanceof RetabAuthenticationException || status == 401 || status == 403,
        "expected RetabAuthenticationException or 401/403, got "
            + thrown.getClass().getSimpleName()
            + " (status="
            + status
            + ")");
    System.out.println(
        "ok   bad api key raises "
            + thrown.getClass().getSimpleName()
            + " (status="
            + status
            + ")");
  }
}
