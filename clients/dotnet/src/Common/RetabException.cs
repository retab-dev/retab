// @oagen-ignore-file
// Hand-maintained — exception hierarchy mirroring the Python / Node SDKs.

using System;

namespace Retab
{
    /// <summary>Base Retab API exception.</summary>
    public class RetabException : Exception
    {
        /// <summary>The HTTP status code returned by the API.</summary>
        public int StatusCode { get; }

        /// <summary>The raw response body returned by the API.</summary>
        public string ResponseBody { get; }

        public RetabException(string message, int statusCode, string responseBody) : base(message)
        {
            this.StatusCode = statusCode;
            this.ResponseBody = responseBody;
        }

        /// <summary>Build the appropriate RetabException subclass for a given HTTP status.</summary>
        public static RetabException From(int statusCode, string body)
        {
            return statusCode switch
            {
                400 => new RetabBadRequestException(body),
                401 => new RetabAuthenticationException(body),
                403 => new RetabPermissionException(body),
                404 => new RetabNotFoundException(body),
                409 => new RetabConflictException(body),
                422 => new RetabValidationException(body),
                429 => new RetabRateLimitException(body),
                _ when statusCode >= 500 => new RetabServerException(statusCode, body),
                _ => new RetabException($"Retab API error ({statusCode})", statusCode, body),
            };
        }
    }

    public class RetabBadRequestException : RetabException
    { public RetabBadRequestException(string body) : base("Retab bad request", 400, body) { } }
    public class RetabAuthenticationException : RetabException
    { public RetabAuthenticationException(string body) : base("Retab authentication failed", 401, body) { } }
    public class RetabPermissionException : RetabException
    { public RetabPermissionException(string body) : base("Retab request not permitted", 403, body) { } }
    public class RetabNotFoundException : RetabException
    { public RetabNotFoundException(string body) : base("Retab resource not found", 404, body) { } }
    public class RetabConflictException : RetabException
    { public RetabConflictException(string body) : base("Retab conflict", 409, body) { } }
    public class RetabValidationException : RetabException
    { public RetabValidationException(string body) : base("Retab validation error", 422, body) { } }
    public class RetabRateLimitException : RetabException
    { public RetabRateLimitException(string body) : base("Retab rate limit exceeded", 429, body) { } }
    public class RetabServerException : RetabException
    { public RetabServerException(int status, string body) : base($"Retab server error ({status})", status, body) { } }
}
