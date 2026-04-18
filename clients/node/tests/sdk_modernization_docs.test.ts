import { describe, expect, test } from "bun:test";

const skillReferencePaths = [
  "/Users/sachaichbiah/Local/retab/open-source/sdk/skills/retab/references/extract.md",
  "/Users/sachaichbiah/Local/retab/open-source/sdk/skills/retab/references/parse.md",
  "/Users/sachaichbiah/Local/retab/open-source/sdk/skills/retab/references/edit.md",
  "/Users/sachaichbiah/Local/retab/open-source/sdk/skills/retab/references/classify.md",
  "/Users/sachaichbiah/Local/retab/open-source/sdk/skills/retab/references/split.md",
  "/Users/sachaichbiah/Local/retab/open-source/sdk/skills/retab/references/workflows.md",
];

const cookbookPaths = [
  "/Users/sachaichbiah/Local/retab/open-source/sdk/cookbook/typescript/documents/extract_api.ts",
  "/Users/sachaichbiah/Local/retab/open-source/sdk/cookbook/typescript/documents/extract_api_from_buffer.ts",
  "/Users/sachaichbiah/Local/retab/open-source/sdk/cookbook/typescript/documents/create_messages.ts",
];

describe("open-source sdk docs modernization", () => {
  test("keeps skill references on resource-style routes and methods", async () => {
    for (const path of skillReferencePaths) {
      const source = await Bun.file(path).text();

      expect(source).not.toContain("/v1/documents/");
      expect(source).not.toContain("result.form_data");
      expect(source).not.toContain("result.filled_document");
      expect(source).not.toContain("choices[0].message.parsed");
    }
  });

  test("keeps cookbook typescript document examples on current sdk methods", async () => {
    for (const path of cookbookPaths) {
      const source = await Bun.file(path).text();

      expect(source).not.toContain("client.documents.");
      expect(source).not.toContain("/v1/documents/");
    }

    const extractExample = await Bun.file(cookbookPaths[0]).text();
    const extractBufferExample = await Bun.file(cookbookPaths[1]).text();
    const parseExample = await Bun.file(cookbookPaths[2]).text();

    expect(extractExample).toContain("client.extractions.create");
    expect(extractExample).toContain("response.output");
    expect(extractBufferExample).toContain("client.extractions.create");
    expect(extractBufferExample).toContain("response.output");
    expect(parseExample).toContain("client.parses.create");
    expect(parseExample).toContain("result.output.pages");
  });

  test("documents workflow waiting and step-inspection helpers in the skill reference", async () => {
    const skill = await Bun.file(
      "/Users/sachaichbiah/Local/retab/open-source/sdk/skills/retab/SKILL.md",
    ).text();
    const workflowRef = await Bun.file(
      "/Users/sachaichbiah/Local/retab/open-source/sdk/skills/retab/references/workflows.md",
    ).text();

    expect(skill).toContain("waiting_for_human");
    expect(skill).toContain("step outputs");

    expect(workflowRef).toContain("createAndWait");
    expect(workflowRef).toContain("wait_for_completion");
    expect(workflowRef).toContain("steps.get");
    expect(workflowRef).toContain("getAll");
    expect(workflowRef).toContain("waiting_for_human");
  });
});
