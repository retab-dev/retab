import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.retab.RetabClient;
import com.retab.models.AssertionSpec;
import com.retab.models.Condition;
import com.retab.models.ExperimentDocumentCaptureRequest;
import com.retab.models.ExistCondition;
import com.retab.models.OutputTarget;
import com.retab.models.PublicHandlePayload;
import com.retab.models.RunStepWorkflowTestSource;
import com.retab.models.WorkflowExperiment;
import com.retab.models.WorkflowRun;
import com.retab.models.WorkflowRunStep;
import com.retab.models.WorkflowTest;
import com.retab.models.WorkflowTestBlockTarget;
import com.retab.types.CreateExperimentRequestNConsensus;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Base64;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public final class Driver {
  static final String API_KEY = "sk_retab_khCajUA8OAo0sFVw89g4CoDW0n9Rv6PBxFdY7OL7XV4";
  static final String BASE_URL = "http://localhost:4000";
  static final String WORKFLOW_ID = "workflow_ClQgW5YKOxUOjDyz2Tt_3";
  static final String START_BLOCK = "block_cKT5b6HgBfjv-XAlXeHJ1";
  static final String PDF =
      "/Users/sachaichbiah/Local/retab/open-source/sdk/assets/docs/Insurance-StateFarm/demand_letter_sample2.pdf";

  static String lifecycleStatus(ObjectMapper om, Object lifecycle) throws Exception {
    if (lifecycle == null) return "null";
    JsonNode n = om.valueToTree(lifecycle);
    return n.has("status") ? n.get("status").asText() : n.toString();
  }

  public static void main(String[] args) throws Exception {
    RetabClient client = new RetabClient(API_KEY, BASE_URL);
    ObjectMapper om = client.getObjectMapper();
    int failures = 0;

    // ---------- 1. CREATE RUN ----------
    System.out.println("=== 1. Create workflow run ===");
    byte[] bytes = Files.readAllBytes(Path.of(PDF));
    String dataUrl = "data:application/pdf;base64," + Base64.getEncoder().encodeToString(bytes);
    Map<String, Object> doc = new LinkedHashMap<>();
    doc.put("filename", "demand_letter_sample2.pdf");
    doc.put("url", dataUrl);
    doc.put("mime_type", "application/pdf");
    Map<String, Object> documents = Map.of(START_BLOCK, doc);

    WorkflowRun run = client.workflows().runs().create(WORKFLOW_ID, documents, null, "production");
    String runId = run.getId();
    System.out.println("created run: " + runId + " status=" + lifecycleStatus(om, run.getLifecycle()));

    // ---------- 2. POLL RUN ----------
    System.out.println("\n=== 2. Poll run to terminal ===");
    String status = lifecycleStatus(om, run.getLifecycle());
    long deadline = System.currentTimeMillis() + 180_000;
    while (!List.of("completed", "error", "cancelled", "awaiting_review").contains(status)
        && System.currentTimeMillis() < deadline) {
      Thread.sleep(3000);
      run = client.workflows().runs().get(runId);
      String s = lifecycleStatus(om, run.getLifecycle());
      if (!s.equals(status)) {
        status = s;
        System.out.println("  -> " + status);
      }
    }
    System.out.println("final run status: " + status);
    if (!status.equals("completed")) {
      System.out.println("WARN: run did not complete cleanly");
    }

    // ---------- 3. STEPS + HANDLE OUTPUTS ----------
    System.out.println("\n=== 3. List steps ===");
    List<WorkflowRunStep> steps =
        client.workflows().steps().list(runId, null, null, null, null, null, null, 100L);
    System.out.println("steps returned: " + (steps == null ? "null" : steps.size()));

    int handlesFromList = 0;
    for (WorkflowRunStep s : steps) {
      Map<String, PublicHandlePayload> ho = s.getHandleOutputs();
      handlesFromList += (ho == null ? 0 : ho.size());
    }
    System.out.println("handle_outputs populated via list(): " + handlesFromList + " (expected 0)");

    System.out.println("\n=== 3b. GET each step, inspect handle_outputs JSON ===");
    int totalJsonHandles = 0;
    String exampleExtractStepId = null;
    String exampleExtractBlockId = null;
    String exampleHandleId = null;
    for (WorkflowRunStep listed : steps) {
      WorkflowRunStep full = client.workflows().steps().get(listed.getStepId(), runId);
      String st = lifecycleStatus(om, full.getLifecycle());
      Map<String, PublicHandlePayload> ho = full.getHandleOutputs();
      int n = (ho == null ? 0 : ho.size());
      System.out.printf(
          "  %-32s %-14s %-11s handles=%d%n",
          full.getBlockId(), full.getBlockType(), st, n);
      if (ho == null) continue;
      for (Map.Entry<String, PublicHandlePayload> e : ho.entrySet()) {
        PublicHandlePayload p = e.getValue();
        Object data = p.getData();
        String type = String.valueOf(p.getType());
        // Clean-JSON check: a "json" payload's data must be a real object/array,
        // never a stringified blob.
        if ("json".equalsIgnoreCase(type) || (data != null)) {
          if (data instanceof String) {
            String ds = ((String) data).trim();
            if (ds.startsWith("{") || ds.startsWith("[")) {
              System.out.println("    !! DIRTY: handle '" + e.getKey()
                  + "' data is a STRINGIFIED json blob");
              failures++;
            }
          }
          if (data != null && !(data instanceof String)) {
            totalJsonHandles++;
            if (exampleHandleId == null && "extract".equals(String.valueOf(full.getBlockType()))) {
              exampleExtractStepId = full.getStepId();
              exampleExtractBlockId = full.getBlockId();
              exampleHandleId = e.getKey();
              System.out.println("    sample clean JSON from handle '" + e.getKey() + "':");
              String pretty = om.writerWithDefaultPrettyPrinter().writeValueAsString(data);
              for (String line : pretty.split("\n")) System.out.println("      " + line);
            }
          }
        }
      }
    }
    System.out.println("structured-json handles found: " + totalJsonHandles);

    // ---------- 4. EXPERIMENTS ----------
    System.out.println("\n=== 4. Experiments ===");
    String expBlock = exampleExtractBlockId != null ? exampleExtractBlockId : "block_HKX5W5phPuuVlTQT1YZiF";
    // n_consensus is an integer enum; the SDK now serializes VALUE_3 as the JSON
    // number 3 (not the string "3"), so the API accepts it instead of 422-ing.
    // Capture the document from the run+step we just executed.
    String capStep = exampleExtractStepId != null ? exampleExtractStepId : runId + "_" + expBlock;
    List<ExperimentDocumentCaptureRequest> captures =
        List.of(new ExperimentDocumentCaptureRequest(runId, capStep));
    WorkflowExperiment exp =
        client.workflows().experiments().create(
            WORKFLOW_ID, expBlock, captures, null,
            CreateExperimentRequestNConsensus.VALUE_3, "java-sdk-smoke-exp", null);
    System.out.println("created experiment: " + exp.getId() + " name=" + exp.getName()
        + " block=" + exp.getBlockId() + " n_consensus=" + exp.getNConsensus());
    WorkflowExperiment expGet = client.workflows().experiments().get(exp.getId());
    System.out.println("get experiment: " + expGet.getId() + " status=" + expGet.getStatus());
    List<WorkflowExperiment> expList =
        client.workflows().experiments().list(WORKFLOW_ID, null, null, 5L, null);
    System.out.println("list experiments (limit 5): " + expList.size());

    // ---------- 5. TESTS ----------
    System.out.println("\n=== 5. Tests ===");
    String testBlock = exampleExtractBlockId != null ? exampleExtractBlockId : "block_HKX5W5phPuuVlTQT1YZiF";
    String testStep = exampleExtractStepId != null ? exampleExtractStepId : runId + "_" + testBlock;
    String handleId = exampleHandleId != null ? exampleHandleId : "output-json-0";

    WorkflowTestBlockTarget target = new WorkflowTestBlockTarget("block", testBlock);
    RunStepWorkflowTestSource source = new RunStepWorkflowTestSource("run_step", runId, testStep);
    Condition cond = new ExistCondition("exists");
    AssertionSpec assertion =
        new AssertionSpec(null, new OutputTarget(handleId, null), cond, "handle exists");

    WorkflowTest test =
        client.workflows().tests().create(WORKFLOW_ID, target, source, "java-sdk-smoke-test", assertion);
    System.out.println("created test: " + test.getId() + " name=" + test.getName()
        + " target_block=" + test.getTarget().getBlockId()
        + " validation_status=" + test.getValidationStatus());
    WorkflowTest testGet = client.workflows().tests().get(test.getId());
    System.out.println("get test: " + testGet.getId() + " assertion_label=" + testGet.getAssertion().getLabel());
    List<WorkflowTest> testList =
        client.workflows().tests().list(WORKFLOW_ID, null, null, null, 5L, null);
    System.out.println("list tests (limit 5): " + testList.size());

    System.out.println("\n=== SUMMARY ===");
    System.out.println("run: " + runId + " (" + status + ")");
    System.out.println("structured json handles: " + totalJsonHandles);
    System.out.println("experiment: " + exp.getId());
    System.out.println("test: " + test.getId());
    System.out.println("dirty-json failures: " + failures);
    System.out.println(failures == 0 ? "RESULT: OK" : "RESULT: FAIL");
  }
}
