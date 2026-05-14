package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	retab "github.com/retab-dev/retab/clients/go"
	"github.com/spf13/cobra"
)

// fileDownloadClient is used for direct downloads from signed storage URLs.
// The timeout is generous (10 min for large files) but bounded so a wedged
// server doesn't hang the CLI in unattended scripts. Ctrl-C still works via
// the request context.
var fileDownloadClient = &http.Client{
	Timeout: 10 * time.Minute,
}

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Manage files",
}

var filesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List files",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		params := retab.ListFilesParams{ListParams: collectListParams(cmd)}
		if v, _ := cmd.Flags().GetString("mime-type"); v != "" {
			params.MIMEType = v
		}
		if v, _ := cmd.Flags().GetString("sort-by"); v != "" {
			params.SortBy = v
		}
		result, err := client.Files.List(ctx, &params)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesGetCmd = &cobra.Command{
	Use:   "get <file-id>",
	Short: "Get a file by id",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Files.Get(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesUploadCmd = &cobra.Command{
	Use:   "upload <path>",
	Short: "Upload a local file",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Files.Upload(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesDownloadLinkCmd = &cobra.Command{
	Use:   "download-link <file-id>",
	Short: "Get a download link for a file",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		result, err := client.Files.GetDownloadLink(ctx, args[0])
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesDownloadCmd = &cobra.Command{
	Use:   "download <file-id>",
	Short: "Download a file to a local path (- for stdout)",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		link, err := client.Files.GetDownloadLink(ctx, args[0])
		if err != nil {
			return err
		}
		out, _ := cmd.Flags().GetString("output")
		if out == "" {
			out = link.Filename
		}
		if out == "" {
			out = args[0]
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, link.DownloadURL, nil)
		if err != nil {
			return err
		}
		resp, err := fileDownloadClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("download failed: %d %s", resp.StatusCode, string(body))
		}
		var sink io.Writer
		if out == "-" {
			sink = os.Stdout
		} else {
			f, err := os.Create(out)
			if err != nil {
				return err
			}
			defer f.Close()
			sink = f
		}
		_, err = io.Copy(sink, resp.Body)
		return err
	}),
}

var filesCreateUploadCmd = &cobra.Command{
	Use:   "create-upload",
	Short: "Reserve a direct-to-storage upload session",
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		filename, _ := cmd.Flags().GetString("filename")
		contentType, _ := cmd.Flags().GetString("content-type")
		sizeStr, _ := cmd.Flags().GetString("size-bytes")
		sha256Hash, _ := cmd.Flags().GetString("sha256")
		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid --size-bytes: %w", err)
		}
		result, err := client.Files.CreateUpload(ctx, retab.PrepareUploadRequest{
			Filename:    filename,
			ContentType: contentType,
			SizeBytes:   size,
			SHA256:      sha256Hash,
		})
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

var filesCompleteUploadCmd = &cobra.Command{
	Use:   "complete-upload <file-id>",
	Short: "Mark a direct upload as finished",
	Args:  cobra.ExactArgs(1),
	RunE: runE(func(cmd *cobra.Command, args []string) error {
		client, err := newClient(cmd)
		if err != nil {
			return err
		}
		ctx, cancel := ctxFor(cmd)
		defer cancel()
		sha256Hash, _ := cmd.Flags().GetString("sha256")
		result, err := client.Files.CompleteUpload(ctx, args[0], sha256Hash)
		if err != nil {
			return err
		}
		return printJSON(result)
	}),
}

func init() {
	addListFlags(filesListCmd, false)
	filesListCmd.Flags().String("mime-type", "", "filter by MIME type")
	filesListCmd.Flags().String("sort-by", "", "sort field (default: created_at)")

	filesDownloadCmd.Flags().StringP("output", "o", "", "output path (default: server filename, - for stdout)")

	filesCreateUploadCmd.Flags().String("filename", "", "filename (required)")
	filesCreateUploadCmd.Flags().String("content-type", "", "content type (required)")
	filesCreateUploadCmd.Flags().String("size-bytes", "0", "file size in bytes (required)")
	filesCreateUploadCmd.Flags().String("sha256", "", "sha256 hex digest (optional)")
	_ = filesCreateUploadCmd.MarkFlagRequired("filename")
	_ = filesCreateUploadCmd.MarkFlagRequired("content-type")
	_ = filesCreateUploadCmd.MarkFlagRequired("size-bytes")

	filesCompleteUploadCmd.Flags().String("sha256", "", "sha256 hex digest (optional)")

	filesCmd.AddCommand(filesListCmd, filesGetCmd, filesUploadCmd, filesDownloadLinkCmd, filesDownloadCmd, filesCreateUploadCmd, filesCompleteUploadCmd)
	rootCmd.AddCommand(filesCmd)
}
