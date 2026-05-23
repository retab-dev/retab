// @oagen-ignore-file
// Hand-maintained — paginated list wrapper returned by every list endpoint.

using System;
using System.Collections.Generic;
using System.Runtime.CompilerServices;
using System.Text.Json.Serialization;
using System.Threading;
using System.Threading.Tasks;

namespace Retab
{
    /// <summary>Paginated list of <typeparamref name="T"/> values returned by list endpoints.</summary>
    public class PaginatedList<T>
    {
        [JsonPropertyName("data")]
        public List<T> Data { get; set; }

        [JsonPropertyName("list_metadata")]
        public ListMetadata ListMetadata { get; set; }

        [JsonPropertyName("object")]
        public string? Object { get; set; }

        internal Func<string, CancellationToken, Task<PaginatedList<T>>>? FetchNextPage { get; set; }

        public bool HasNextPage => !string.IsNullOrEmpty(this.ListMetadata?.After);

        public PaginatedList()
        {
            this.Data = new List<T>();
            this.ListMetadata = new ListMetadata();
        }

        public PaginatedList(
            List<T>? data,
            ListMetadata listMetadata,
            string? objectType = null,
            Func<string, CancellationToken, Task<PaginatedList<T>>>? fetchNextPage = null)
        {
            this.Data = data ?? new List<T>();
            this.ListMetadata = listMetadata ?? new ListMetadata();
            this.Object = objectType;
            this.FetchNextPage = fetchNextPage;
        }

        public async IAsyncEnumerable<T> AutoPagingIterAsync(
            [EnumeratorCancellation] CancellationToken cancellationToken = default)
        {
            var current = this;
            while (current != null)
            {
                if (current.Data != null)
                {
                    foreach (var item in current.Data)
                    {
                        cancellationToken.ThrowIfCancellationRequested();
                        yield return item;
                    }
                }

                var after = current.ListMetadata?.After;
                if (string.IsNullOrEmpty(after) || current.FetchNextPage == null)
                {
                    yield break;
                }

                current = await current.FetchNextPage(after, cancellationToken).ConfigureAwait(false);
            }
        }

        public IAsyncEnumerator<T> GetAsyncEnumerator(CancellationToken cancellationToken = default)
            => this.AutoPagingIterAsync(cancellationToken).GetAsyncEnumerator(cancellationToken);
    }
}
