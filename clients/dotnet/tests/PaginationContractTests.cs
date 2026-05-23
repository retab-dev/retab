// @oagen-ignore-file
// Reflection-driven pagination contract test. Every public Service.ListAsync
// returning Task<PaginatedList<T>> must wire the next-page closure so
// await foreach over the first page follows list_metadata.after.

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

        private static IEnumerable<(string ServiceName, PropertyInfo Accessor, MethodInfo ListAsync, Type ItemType)> DiscoverListMethods()
        {
            var clientType = typeof(RetabClient);
            var accessors = clientType.GetProperties(BindingFlags.Public | BindingFlags.Instance)
                .Where(p => typeof(Service).IsAssignableFrom(p.PropertyType));

            foreach (var accessor in accessors)
            {
                var serviceType = accessor.PropertyType;
                var listAsync = serviceType.GetMethod("ListAsync", BindingFlags.Public | BindingFlags.Instance);
                if (listAsync == null) continue;

                var returnType = listAsync.ReturnType;
                if (!returnType.IsGenericType || returnType.GetGenericTypeDefinition() != typeof(Task<>)) continue;

                var inner = returnType.GenericTypeArguments[0];
                if (!inner.IsGenericType || inner.GetGenericTypeDefinition() != typeof(PaginatedList<>)) continue;

                yield return (accessor.Name, accessor, listAsync, inner.GenericTypeArguments[0]);
            }
        }

        public static IEnumerable<object[]> EveryListService()
        {
            foreach (var (name, _, _, _) in DiscoverListMethods())
            {
                yield return new object[] { name };
            }
        }

        [Fact]
        public void ContractTestEnumeratesEveryListService()
        {
            var discovered = DiscoverListMethods().ToList();
            Assert.True(discovered.Count >= 20, $"Expected to discover 20+ paginated services, found {discovered.Count}.");
        }

        [Theory]
        [MemberData(nameof(EveryListService))]
        public async Task ListAsyncAutoPagesAcrossMultiplePages(string serviceName)
        {
            if (KnownBypass.Contains(serviceName))
            {
                return;
            }

            var match = DiscoverListMethods().First(t => t.ServiceName == serviceName);
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

            var service = match.Accessor.GetValue(client);
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
                .MakeGenericMethod(match.ItemType);
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
