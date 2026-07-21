// @oagen-ignore-file
// Reflection-driven pagination contract test. Every public Service.List*Async
// returning Task<PaginatedList<T>> must wire the next-page closure so
// await foreach over the first page follows list_metadata.after.
//
// "List method" means ListAsync OR any List<Suffix>Async variant. An earlier
// version of the walk called GetMethod("ListAsync") — the exact name only —
// which silently excluded Usage.ListBlocksAsync, Workflows.ListVersionsAsync
// and friends, all of which return Task<PaginatedList<T>> and carry the same
// closure obligation.
//
// Discovery matches on NAME, not return type. A List*Async that does not
// return Task<PaginatedList<T>> must be named in NonCursor with a reason —
// otherwise it fails EveryListMethodIsPaginatedOrExempt. Filtering by return
// type alone would let a genuinely-paginated route that regressed to the
// wrong type disappear from coverage instead of failing.

using System;
using System.Collections.Generic;
using System.Linq;
using System.Net;
using System.Net.Http;
using System.Reflection;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using Xunit;
using Retab;
using RetabClient = Retab.Retab;

namespace Retab.Tests
{
    internal sealed class QueueingHandler : HttpMessageHandler
    {
        private readonly Queue<string> responses;

        public int CallCount { get; private set; }

        public List<Uri> Calls { get; } = new List<Uri>();

        public QueueingHandler(params string[] jsonResponses)
        {
            this.responses = new Queue<string>(jsonResponses);
        }

        protected override Task<HttpResponseMessage> SendAsync(HttpRequestMessage request, CancellationToken cancellationToken)
        {
            this.CallCount++;
            this.Calls.Add(request.RequestUri!);
            var body = this.responses.Count > 0
                ? this.responses.Dequeue()
                : "{\"data\":[],\"list_metadata\":{\"before\":null,\"after\":null}}";
            return Task.FromResult(new HttpResponseMessage(HttpStatusCode.OK)
            {
                Content = new StringContent(body, Encoding.UTF8, "application/json"),
            });
        }
    }

    public class PaginationContractTests
    {
        private static readonly HashSet<string> KnownBypass = new HashSet<string>(StringComparer.Ordinal)
        {
            // None today. Additions here must be documented in the SDK pagination blueprint.
        };

        /// <summary>
        /// List methods that legitimately return something other than
        /// Task&lt;PaginatedList&lt;T&gt;&gt;, so the cursor contract does not apply.
        /// Each entry carries its reason. Stale rows fail
        /// <see cref="NonCursorEntriesAllResolve"/> — they must keep naming a live
        /// method that is still non-paginated.
        /// </summary>
        private static readonly Dictionary<string, string> NonCursor = new Dictionary<string, string>(StringComparer.Ordinal)
        {
            ["Tables"] = "ListAsync returns WorkflowTableListResponse ({tables: [...]}), no cursor envelope",
            ["Secrets"] = "ListAsync returns SecretListResponse, an unpaginated envelope",
            ["Secrets.ListValueAsync"] = "returns SecretValueResponse; 'list' is the API verb for a single value read, not a collection",
            ["Workflows.ListDiffAsync"] = "returns a single WorkflowGraphVersionDiff",
            ["Workflows.Blocks.ListDiffAsync"] = "returns a single WorkflowBlockVersionDiff",
            ["Workflows.Edges.ListDiffAsync"] = "returns a single WorkflowEdgeVersionDiff",
        };

        /// <summary>
        /// The List*Async variants that MUST stay in the contract. If discovery
        /// ever narrows back to the exact name "ListAsync" these drop out
        /// silently and every remaining case still passes.
        /// </summary>
        private static readonly string[] PaginatedListStarMethods = new[]
        {
            "Usage.ListBlocksAsync",
            "Usage.ListPrimitivesAsync",
            "Usage.ListRunsAsync",
            "Workflows.ListVersionsAsync",
            "Workflows.Blocks.ListVersionsAsync",
            "Workflows.Edges.ListVersionsAsync",
        };

        /// <summary>
        /// True for ListAsync and every List&lt;Suffix&gt;Async variant. The
        /// AutoPaging helpers (List*AutoPagingAsync) are the convenience
        /// IAsyncEnumerable wrappers *around* the contract, not routes of their
        /// own, so they are excluded by name.
        /// </summary>
        private static bool IsListMethodName(string name)
        {
            if (!name.StartsWith("List", StringComparison.Ordinal) || !name.EndsWith("Async", StringComparison.Ordinal))
            {
                return false;
            }

            if (name.Contains("AutoPaging", StringComparison.Ordinal))
            {
                return false;
            }

            var middle = name.Substring("List".Length, name.Length - "List".Length - "Async".Length);
            // "" => bare ListAsync; otherwise the suffix must start upper-case
            // so we don't match a hypothetical "ListenAsync".
            return middle.Length == 0 || char.IsUpper(middle[0]);
        }

        /// <summary>
        /// Discovery label, mirroring the cross-language rule: the dotted
        /// accessor path alone for the bare ListAsync, and path + method name
        /// for a List*Async variant.
        /// </summary>
        private static string Label(string servicePath, string methodName)
            => methodName == "ListAsync" ? servicePath : servicePath + "." + methodName;

        private static IEnumerable<(string ServiceName, PropertyInfo[] AccessPath, MethodInfo ListAsync, Type? ItemType)> DiscoverListMethods()
        {
            return DiscoverListMethods(typeof(RetabClient), Array.Empty<PropertyInfo>(), new HashSet<Type>());
        }

        private static IEnumerable<(string ServiceName, PropertyInfo[] AccessPath, MethodInfo ListAsync, Type? ItemType)> DiscoverListMethods(
            Type ownerType,
            PropertyInfo[] path,
            HashSet<Type> visited)
        {
            var accessors = ownerType.GetProperties(BindingFlags.Public | BindingFlags.Instance)
                .Where(p => typeof(Service).IsAssignableFrom(p.PropertyType) && p.GetIndexParameters().Length == 0);
            foreach (var accessor in accessors)
            {
                var serviceType = accessor.PropertyType;
                if (!visited.Add(serviceType)) continue;

                var servicePath = path.Concat(new[] { accessor }).ToArray();
                var serviceName = string.Join(".", servicePath.Select(p => p.Name));

                var listMethods = serviceType.GetMethods(BindingFlags.Public | BindingFlags.Instance)
                    .Where(m => IsListMethodName(m.Name))
                    .OrderBy(m => m.Name, StringComparer.Ordinal);
                foreach (var listAsync in listMethods)
                {
                    yield return (Label(serviceName, listAsync.Name), servicePath, listAsync, PaginatedItemType(listAsync));
                }

                foreach (var nested in DiscoverListMethods(serviceType, servicePath, visited))
                {
                    yield return nested;
                }
            }
        }

        /// <summary>
        /// The T in Task&lt;PaginatedList&lt;T&gt;&gt;, or null when the method
        /// returns some other envelope.
        /// </summary>
        private static Type? PaginatedItemType(MethodInfo method)
        {
            var returnType = method.ReturnType;
            if (!returnType.IsGenericType || returnType.GetGenericTypeDefinition() != typeof(Task<>))
            {
                return null;
            }

            var inner = returnType.GenericTypeArguments[0];
            if (!inner.IsGenericType || inner.GetGenericTypeDefinition() != typeof(PaginatedList<>))
            {
                return null;
            }

            return inner.GenericTypeArguments[0];
        }

        /// <summary>Discovered list methods that must honour the cursor contract.</summary>
        private static IEnumerable<(string ServiceName, PropertyInfo[] AccessPath, MethodInfo ListAsync, Type? ItemType)> ContractListMethods()
            => DiscoverListMethods().Where(t => !NonCursor.ContainsKey(t.ServiceName) && !KnownBypass.Contains(t.ServiceName));

        public static IEnumerable<object[]> EveryListService()
        {
            foreach (var (name, _, _, _) in ContractListMethods())
            {
                yield return new object[] { name };
            }
        }

        [Fact]
        public void ContractTestEnumeratesEveryListService()
        {
            var discovered = ContractListMethods().ToList();
            Assert.True(discovered.Count >= 26, $"Expected to discover 26+ paginated list methods, found {discovered.Count}.");
        }

        [Fact]
        public void DiscoveryCoversListStarVariants()
        {
            var discovered = DiscoverListMethods().ToDictionary(t => t.ServiceName, t => t.ItemType);
            foreach (var name in PaginatedListStarMethods)
            {
                Assert.True(
                    discovered.ContainsKey(name),
                    $"{name} was not discovered — the List*Async discovery walk regressed to matching only the exact name \"ListAsync\"."
                );
                Assert.True(
                    discovered[name] != null,
                    $"{name} no longer returns Task<PaginatedList<T>>; the route is cursor-paginated, so this is a real bug."
                );
            }
        }

        [Fact]
        public void EveryListMethodIsPaginatedOrExempt()
        {
            var offenders = ContractListMethods()
                .Where(t => t.ItemType == null)
                .Select(t => t.ServiceName)
                .OrderBy(n => n, StringComparer.Ordinal)
                .ToList();
            Assert.True(
                offenders.Count == 0,
                $"These list methods do not return Task<PaginatedList<T>> and are not in NonCursor/KnownBypass: {string.Join(", ", offenders)}. "
                + "Either fix the return type (the route is cursor-paginated) or add the method to NonCursor with a documented reason."
            );
        }

        [Fact]
        public void NonCursorEntriesAllResolve()
        {
            var discovered = DiscoverListMethods().ToDictionary(t => t.ServiceName, t => t.ItemType);
            foreach (var entry in NonCursor)
            {
                Assert.True(
                    discovered.ContainsKey(entry.Key),
                    $"NonCursor entry \"{entry.Key}\" no longer matches a list method on the client; remove it."
                );
                Assert.False(string.IsNullOrWhiteSpace(entry.Value), $"NonCursor entry \"{entry.Key}\" needs a reason.");
                Assert.True(
                    discovered[entry.Key] == null,
                    $"{entry.Key} is exempted as non-cursor but now returns Task<PaginatedList<T>> — that's an improvement; drop it from NonCursor so the contract covers it."
                );
            }
        }

        [Theory]
        [MemberData(nameof(EveryListService))]
        public async Task ListAsyncAutoPagesAcrossMultiplePages(string serviceName)
        {
            if (KnownBypass.Contains(serviceName))
            {
                return;
            }

            var match = ContractListMethods().First(t => t.ServiceName == serviceName);
            Assert.True(
                match.ItemType != null,
                $"{serviceName} does not return Task<PaginatedList<T>> and is not in NonCursor — "
                + "either fix the return type or document the exemption with a reason."
            );
            var handler = new QueueingHandler(
                "{\"data\":[{}],\"list_metadata\":{\"before\":null,\"after\":\"cursor-2\"}}",
                "{\"data\":[{}],\"list_metadata\":{\"before\":null,\"after\":null}}"
            );
            var http = new HttpClient(handler) { BaseAddress = new Uri("http://stub.local/") };
            var client = new RetabClient(new RetabOptions
            {
                ApiKey = "test-api-key",
                BaseUrl = new Uri("http://stub.local/"),
                HttpClient = http,
            });

            object? service = client;
            foreach (var accessor in match.AccessPath)
            {
                service = accessor.GetValue(service);
            }
            Assert.NotNull(service);

            var parameters = match.ListAsync.GetParameters();
            var args = new object?[parameters.Length];
            for (int i = 0; i < parameters.Length; i++)
            {
                if (parameters[i].ParameterType == typeof(string) && parameters[i].Name == "httpBearer")
                {
                    args[i] = "test-bearer";
                }
                else if (parameters[i].ParameterType == typeof(CancellationToken))
                {
                    args[i] = CancellationToken.None;
                }
                else if (parameters[i].HasDefaultValue)
                {
                    args[i] = parameters[i].DefaultValue;
                }
                else if (parameters[i].ParameterType.IsValueType)
                {
                    args[i] = Activator.CreateInstance(parameters[i].ParameterType);
                }
                else
                {
                    args[i] = null;
                }
            }

            var taskObj = match.ListAsync.Invoke(service, args)!;
            await ((Task)taskObj);
            var page = taskObj.GetType().GetProperty("Result")!.GetValue(taskObj)!;

            var collect = typeof(PaginationContractTests)
                .GetMethod(nameof(CollectAsync), BindingFlags.NonPublic | BindingFlags.Static)!
                .MakeGenericMethod(match.ItemType!);
            var collectedTask = (Task<List<object?>>)collect.Invoke(null, new[] { page })!;
            var collected = await collectedTask;

            Assert.True(
                handler.CallCount > 1,
                $"{serviceName}.ListAsync only issued {handler.CallCount} HTTP request(s); auto-pagination did not follow list_metadata.after."
            );
            Assert.Equal(2, collected.Count);
            Assert.Contains(handler.Calls, uri => uri.Query.Contains("after=cursor-2"));
            Assert.DoesNotContain(handler.Calls, uri => uri.Query.Contains("before="));
        }

        private static async Task<List<object?>> CollectAsync<T>(PaginatedList<T> page)
        {
            var collected = new List<object?>();
            await foreach (var item in page)
            {
                collected.Add(item);
            }

            return collected;
        }
    }
}
