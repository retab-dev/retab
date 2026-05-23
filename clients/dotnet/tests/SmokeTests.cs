// @oagen-ignore-file
// Compile-time smoke tests: the assertions below don't make real network
// calls; they exist to verify that the generated Retab client surface
// compiles against the hand-maintained MimeData ergonomics.

using System;
using System.Collections.Generic;
using System.IO;
using System.Threading.Tasks;
using Xunit;
using Retab;

public class SmokeTests
{
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
}
