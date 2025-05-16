# Image vs. Text: What Modality to Use for Document Processing?

Hey there, I'm [Sacha](https://x.com/sachaicb), creator and maintainer of UiForm.

I've been working on document processing for a while now, and I've seen a lot of people surprised when i told them that the best way to process emails or excel files is to use vision, and not language. In this article, I'll share the heuristics we use to determine the optimal approach for each kind of document.

When processing documents, selecting the right modality is crucial for efficiency and accuracy. The two primary modalities available today in most large language models are **text** and **vision**, with **audio** emerging but not yet widely applicable. Often, it is not obvious which modality is best suited for a given document. In this article, we provide the heuristics we use to determine the optimal approach.

## **Modalities**

With UiForm, you can easily switch between modalities:

```python
from uiform import UiForm

uiclient = UiForm()

doc_msg = uiclient.documents.create_messages(
    document="contract.pdf",
    modality="text", # or "image", "image+text", "native"
)
```

```

But which modality should you choose for each use case?

## **Document Types**

We consider as documents the following file types:

- **Excel files** (.xls, .xlsx, .ods)
- **Word files** (.doc, .docx, .odt)
- **PowerPoint files** (.ppt, .pptx, .odp)
- **PDF files** (.pdf)

### **Short Documents (<= 3 pages): Use Images**

Short documents, typically three pages or less, are often scans of images and contain information-dense visuals, making image processing the best modality.

Excel files and tabular data in general are an interesting case-study. They contains a lot of spatial information, that is much better captured by vision language models (VLMs) than text-based models. (The predecessor of Excel was literally named "VisiCalc".)

### **Long Documents (> 3 pages): Use Text**

Long documents, typically over three pages, are text-heavy and structured with paragraphs, headings, and lists. They are often contracts, reports, and documents where the text is the main content. Text-based processing excels here, offering scalability, speed, without sacrificing accuracy.

Using images for long documents can be challenging because high-quality images take up a lot of space and slow things down. Lowering the image resolution to save space can reduce quality and lead to errors. Text processing avoids these issues, giving you quick and direct access to the content without the hassle.

This makes text processing the go-to option for longer documents where understanding the content is more important than preserving its look.

## **Images**

When working with standalone images (e.g., .jpg, .png, .bmp, etc.), using the vision modality is the natural choice.

## **Emails**

Emails are complex because they consist of both a body and attachments, each requiring a different approach to ensure accurate processing and data extraction.

### **Email Body**

The body of an email is more than just plain text—it often includes HTML formatting, embedded images, and links. HTML emails can contain a variety of styled elements, such as tables, fonts, and inline images that help present information in a structured way. Processing email bodies as images can capture their full visual representation, ensuring that layouts, branding elements, and inline attachments are accurately preserved. However, for emails with simpler content, text processing can be more efficient for extracting key information such as dates, sender details, and action items.

It's also important to consider email threading, where multiple messages are grouped together in a conversation. Efficient processing should be able to separate these threads and extract relevant content without losing context.

For this reason, in UiForm, we put in the context window of the LLM both the text content of the body, and the images rendering the html content of the body.

### **Email Attachments**

Attachments come in many different formats—PDFs, Excel files, Word documents, and even images. Each type should be processed using the most suitable method:

- **Text-based attachments** (e.g., Word, PDF with selectable text, CSV) should be handled using text processing for easy data extraction.
- **Image-based attachments** (e.g., scanned PDFs, JPEGs, PNGs) should be processed with OCR to extract any embedded text.
- **Excel files** should be processed either as images for visual fidelity or as text when data needs to be structured for analysis.

Handling attachments efficiently ensures that all relevant information within an email thread is accurately captured and categorised, enhancing data retrieval and workflow automation.

## **Web Pages**

Web pages can be saved as a single file format (e.g., `.mhtml`), which closely resembles an email structure. Similar to emails, web pages contain both text and embedded images, making image processing a great choice.

## **A Hybrid Approach: Image + Text Modality**
In some cases, combining both modalities offers the best results. For example, on complex documents, the LLM can benefit from the image to understand the layout, but struggles to extract the text precisely. In this case, having the image and the text in the context window of the LLM is a great way to get the best of both worlds.

```python
from uiform import UiForm

uiclient = UiForm()

doc_msg = uiclient.documents.create_messages(
    document="contract.pdf",
    modality="image+text",
)
```

## Supported File Types by Modality

```python
# Text-Based Files
TEXT_TYPES = Literal[
".txt", ".csv", ".tsv", ".md", ".log", ".html", ".htm", ".xml", ".json", ".yaml", ".yml",
".rtf", ".ini", ".conf", ".cfg", ".nfo", ".srt", ".sql", ".sh", ".bat", ".ps1", ".js", ".jsx",
".ts", ".tsx", ".py", ".java", ".c", ".cpp", ".cs", ".rb", ".php", ".swift", ".kt", ".go", ".rs",
".pl", ".r", ".m", ".scala"
]

# Image-Based Files
IMAGE_TYPES = Literal[".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff"]

# Document Files (image or text depending on the length)
EXCEL_TYPES = Literal[".xls", ".xlsx", ".ods"]
WORD_TYPES = Literal[".doc", ".docx", ".odt"]
PPT_TYPES = Literal[".ppt", ".pptx", ".odp"]
PDF_TYPES = Literal[".pdf"]

# Email Files
EMAIL_TYPES = Literal[".eml", ".msg"]  # MIME files containing other MIME files

# Web Files
WEB_TYPES = Literal[".mhtml"]
```

By following these heuristics and choosing the right modality, you can achieve the best accuracy and efficiency in document processing.

## Conclusion

Choosing the right way to process documents—whether using text or vision—can make a huge difference in how accurate and efficient your results are. By following the simple guidelines we’ve shared, you can easily decide which approach works best for different types of documents. And for those tricky cases, combining both can give you the best of both worlds.

At UiForm, we’ve made it easy to switch between modalities so you don’t have to worry about choosing the wrong one. Whether you’re working with emails, spreadsheets, or contracts, using the right method will help you get the most out of your data while saving time and effort.

Don't hesitate to reach out on [X](https://x.com/sachaicb) or [Discord](https://discord.gg/vc5tWRPqag) if you have any questions or feedback!