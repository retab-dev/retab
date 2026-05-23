<!-- @oagen-ignore-file -->
# Retab .NET SDK

Official .NET (C#) SDK for the [Retab API](https://retab.com).

```csharp
using Retab;

var client = new Retab("YOUR_API_KEY");

// MimeData accepts FileInfo, byte[], and Uri implicitly:
var extraction = await client.Extractions.CreateAsync(
    new FileInfo("/path/to/invoice.pdf"),
    new Dictionary<string, object>(),
    model: "retab-small"
);
```

## Installation

```
dotnet add package Retab
```

## Building from source

```
dotnet build
dotnet test --no-run
```

The SDK targets `net8.0` with C# 12 nullable reference types enabled. Treat
warnings as errors in CI to catch contract drift across regenerations.
