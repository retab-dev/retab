// @oagen-ignore-file
// Compile-time smoke tests: the assertions below don't make real network
// calls; they exist to verify that the generated Retab client surface
// compiles against the hand-maintained MimeData ergonomics.

using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading;
using System.Threading.Tasks;
using Xunit;
using Retab;

public class SmokeTests
{
    private const string InvoiceWorkflowYaml = """
apiVersion: workflows.retab.com/v1alpha2
kind: Workflow
metadata:
  id: wf_invoice_validation
  name: Invoice Validation Workflow
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
    validate_total:
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
            computed_total:
              type: number
            is_valid:
              type: boolean
          required:
            - invoice_id
            - subtotal
            - computed_total
            - is_valid
        code: |
          from models import Input, Output

          def transform(input_data: Input) -> Output:
              subtotal = sum(item.amount for item in input_data.line_items)
              computed_total = round(subtotal + subtotal * input_data.tax_rate, 2)
              return Output(
                  invoice_id=input_data.invoice_id,
                  subtotal=subtotal,
                  computed_total=computed_total,
                  is_valid=abs(computed_total - input_data.stated_total) <= 0.01,
              )
  edges:
    - from:
        block: start
        handle: output-json-0
      to:
        block: validate_total
        handle: input-json-0
""";

    private sealed class CapturingHandler : HttpMessageHandler
    {
        private readonly string responseBody;

        public HttpRequestMessage? Request { get; private set; }

        public CapturingHandler(string responseBody = "{}")
        {
            this.responseBody = responseBody;
        }

        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            this.Request = request;
            return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(this.responseBody, Encoding.UTF8, "application/json"),
            });
        }
    }

    [Fact]
    public void ClientConstructs()
    {
        var client = new global::Retab.Retab("test-api-key");
        Assert.NotNull(client);
    }

    [Fact]
    public void MimeDataFromFileInfoImplicitConverts()
    {
        // Implicit conversion compiles. Don't actually read the file.
        var info = new FileInfo("/tmp/nonexistent.pdf");
        // The line below is the actual compile-time interop check:
        Action verify = () => { MimeData m = info; _ = m; };
        Assert.NotNull(verify);
    }

    [Fact]
    public void MimeDataFromBytesImplicitConverts()
    {
        byte[] bytes = new byte[] { 0x25, 0x50, 0x44, 0x46 }; // %PDF magic bytes
        MimeData m = bytes;
        Assert.NotNull(m);
        Assert.StartsWith("data:application/pdf;base64,", m.Url);
    }

    [Fact]
    public void MimeDataFromUrlImplicitConverts()
    {
        MimeData m = new Uri("https://example.com/doc.pdf");
        Assert.Equal("doc.pdf", m.Filename);
        Assert.Equal("https://example.com/doc.pdf", m.Url);
    }

    [Fact]
    public void MimeDataFromDataUrlPassthrough()
    {
        var m = MimeData.FromDataUrl("data:application/pdf;base64,JVBERi0=", "passport.pdf");
        Assert.Equal("passport.pdf", m.Filename);
        Assert.StartsWith("data:application/pdf;base64,", m.Url);
    }

    [Fact]
    public void WorkflowRunDocumentsAcceptMimeDataAndFileRefPerBlock()
    {
        var options = new WorkflowRunsCreateOptions
        {
            WorkflowId = "wrk_123",
            Documents = new Dictionary<string, WorkflowRunDocumentInput>
            {
                ["start_pdf"] = MimeData.FromDataUrl("data:application/pdf;base64,JVBERi0=", "passport.pdf"),
                ["start_existing"] = new FileRef
                {
                    Id = "file_123",
                    Filename = "stored.pdf",
                    MimeType = "application/pdf",
                },
            },
        };

        var json = JsonSerializer.Serialize(options, Retab.Retab.JsonOptions);

        Assert.Contains("\"start_pdf\":{\"filename\":\"passport.pdf\",\"url\":\"data:application/pdf;base64,JVBERi0=\"}", json);
        Assert.Contains("\"start_existing\":{\"id\":\"file_123\",\"filename\":\"stored.pdf\",\"mime_type\":\"application/pdf\"}", json);
    }

    [Fact]
    public async Task ClientSendsApiKeyHeaderByDefault()
    {
        var handler = new CapturingHandler();
        var client = new global::Retab.Retab(new RetabOptions
        {
            ApiKey = "test-api-key",
            BaseUrl = new Uri("http://stub.local"),
            HttpClient = new HttpClient(handler),
        });

        await client.MakeAPIRequest<Dictionary<string, object>>(
            new RetabRequest { Method = HttpMethod.Get, Path = "/v1/workflows" },
            CancellationToken.None
        );

        Assert.NotNull(handler.Request);
        Assert.Equal("Bearer test-api-key", handler.Request!.Headers.Authorization!.ToString());
    }

    [Fact]
    public async Task RequestOptionsApiKeyOverridesClientApiKey()
    {
        var handler = new CapturingHandler();
        var client = new global::Retab.Retab(new RetabOptions
        {
            ApiKey = "client-api-key",
            BaseUrl = new Uri("http://stub.local"),
            HttpClient = new HttpClient(handler),
        });

        await client.MakeAPIRequest<Dictionary<string, object>>(
            new RetabRequest
            {
                Method = HttpMethod.Get,
                Path = "/v1/workflows",
                RequestOptions = new RequestOptions { ApiKey = "override-api-key" },
            },
            CancellationToken.None
        );

        Assert.NotNull(handler.Request);
        Assert.Equal("Bearer override-api-key", handler.Request!.Headers.Authorization!.ToString());
    }

    [Fact]
    public void BaseUrlMayIncludeVersionPrefix()
    {
        var client = new global::Retab.Retab(new RetabOptions
        {
            ApiKey = "test-api-key",
            BaseUrl = new Uri("http://localhost:4000/v1"),
        });

        var uri = client.BuildRequestUri(new RetabRequest { Path = "/v1/workflows" });

        Assert.Equal("http://localhost:4000/v1/workflows", uri.ToString());
    }

    [Fact]
    public void BodyMethodOptionsDoNotBecomeQueryParams()
    {
        var client = new global::Retab.Retab(new RetabOptions
        {
            ApiKey = "test-api-key",
            BaseUrl = new Uri("http://stub.local"),
        });

        var uri = client.BuildRequestUri(new RetabRequest
        {
            Method = HttpMethod.Post,
            Path = "/v1/workflows/spec/apply",
            Options = new WorkflowsApplyOptions { YamlDefinition = InvoiceWorkflowYaml },
            ExtraQuery = new Dictionary<string, string> { ["trace"] = "1" },
        });

        Assert.Equal("http://stub.local/v1/workflows/spec/apply?trace=1", uri.ToString());

        var workflowRunUri = client.BuildRequestUri(new RetabRequest
        {
            Method = HttpMethod.Post,
            Path = "/v1/workflows/runs",
            Options = new WorkflowRunsCreateOptions
            {
                WorkflowId = "wf_123",
                JsonInputs = new Dictionary<string, object>
                {
                    ["start"] = new Dictionary<string, object>
                    {
                        ["invoice_id"] = "inv_sdk_smoke",
                        ["line_items"] = new[]
                        {
                            new Dictionary<string, object> { ["description"] = "warehouse handling", ["amount"] = 120 },
                            new Dictionary<string, object> { ["description"] = "local delivery", ["amount"] = 80 },
                        },
                        ["tax_rate"] = 0.2,
                        ["stated_total"] = 240,
                    },
                },
            },
        });

        Assert.Equal("http://stub.local/v1/workflows/runs", workflowRunUri.ToString());
    }

    [Fact]
    public async Task ResponseDiscriminatorConvertersRun()
    {
        var handler = new CapturingHandler(
            "{" +
            "\"id\":\"run_123\"," +
            "\"workflow\":{\"workflow_id\":\"wrk_123\",\"version_id\":\"ver_123\"}," +
            "\"trigger\":{\"type\":\"api\"}," +
            "\"lifecycle\":{\"status\":\"completed\"}" +
            "}"
        );
        var client = new global::Retab.Retab(new RetabOptions
        {
            ApiKey = "test-api-key",
            BaseUrl = new Uri("http://stub.local"),
            HttpClient = new HttpClient(handler),
        });

        var run = await client.MakeAPIRequest<WorkflowRun>(
            new RetabRequest { Method = HttpMethod.Get, Path = "/v1/workflows/runs/run_123" },
            CancellationToken.None
        );

        var trigger = Assert.IsType<TriggerInfo>(run.Trigger);
        Assert.Equal(TriggerInfoType.Api, trigger.Type);
        Assert.IsType<CompletedBlockExecutionLifecycle>(run.Lifecycle);
    }

    [Fact]
    public async Task UnionResponseDecodesNonFirstVariantLosslessly()
    {
        // Regression for the lossy discriminated-union RESPONSE collapse:
        // GET /v1/workflows/artifacts/{id} is an 11-variant union keyed on
        // "operation". Upstream codegen collapsed the response to the FIRST
        // variant (ExtractionWorkflowArtifact), dropping every other variant's
        // fields. A classification artifact must decode to the classification
        // variant (NOT extraction) with its variant-specific fields intact, and
        // an unmodeled field must survive a deserialize -> serialize round-trip
        // via [JsonExtensionData].
        var handler = new CapturingHandler(
            "{" +
            "\"operation\":\"classification\"," +
            "\"id\":\"clss_123\"," +
            "\"model\":\"retab-small\"," +
            "\"future_field\":\"keep-me\"" +
            "}"
        );
        var client = new global::Retab.Retab(new RetabOptions
        {
            ApiKey = "test-api-key",
            BaseUrl = new Uri("http://stub.local"),
            HttpClient = new HttpClient(handler),
        });

        var artifact = await client.Workflows.Artifacts.GetAsync("clss_123");

        // Decoded to the correct (non-first) variant, not ExtractionWorkflowArtifact.
        var classification = Assert.IsType<ClassificationWorkflowArtifact>(artifact);
        Assert.Equal("clss_123", classification.Id);
        Assert.Equal("retab-small", classification.Model);

        // The unmodeled wire field survived on the extension-data container, so a
        // round-trip back to JSON preserves it.
        var roundTripped = Newtonsoft.Json.JsonConvert.SerializeObject(classification);
        Assert.Contains("future_field", roundTripped);
        Assert.Contains("keep-me", roundTripped);
    }
}
