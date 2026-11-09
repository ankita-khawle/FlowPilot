# FlowPilot 🚀

**Upload a document. Let AI do the reading. Get the insights without the boring bits.**

FlowPilot is an AI-powered document intelligence platform that uses Large Language Models (LLMs) and Retrieval-Augmented Generation (RAG) to analyse uploaded documents, extract meaningful information, and generate contextual summaries in a quirky, engaging tone.

Instead of reading lengthy documents from beginning to end, users can upload a file and let FlowPilot identify important information, understand context, and present the key takeaways in a more enjoyable and accessible way.

## ✨ What FlowPilot Does

- **Document Upload:** Upload PDF, DOCX, and TXT documents.
- **AI-Powered Analysis:** Use LLMs to understand document content and identify important information.
- **RAG-Based Context Retrieval:** Retrieve relevant document sections to provide context-aware and grounded responses.
- **Quirky AI Summaries:** Generate engaging summaries that explain complex or lengthy content in a fun, conversational tone.
- **Structured Data Extraction:** Extract important entities, dates, figures, decisions, and other relevant information.
- **Source Citations:** Link generated insights to their original document sections or page references for traceability.
- **Customisable Summaries:** Choose the tone and summary length based on your preferences.

## 🧠 How It Works

1. **Upload:** Select a document to analyse.
2. **Extract:** Extract text and relevant content from the uploaded file.
3. **Index:** Split the document into meaningful chunks and generate embeddings for retrieval.
4. **Retrieve:** Use RAG to identify relevant sections based on the analysis task.
5. **Analyse:** Send the retrieved context to an LLM for summarisation and structured data extraction.
6. **Generate:** Produce a quirky, easy-to-understand summary with supporting citations.
7. **Review:** View the summary, extracted information, and source references.

## 🛠️ Tech Stack

| Layer | Technology |
|---|---|
| Frontend | React, TypeScript, Vite |
| Backend | Go |
| AI | Large Language Models (LLMs) |
| Retrieval | RAG, embeddings, vector search |
| Document Processing | PDF, DOCX, TXT extraction |
| API | REST |
| Future storage | PostgreSQL, vector database |

## 🏗️ Project Structure

```text
FlowPilot/
├── backend/
│   ├── cmd/
│   └── internal/
│       ├── httpapi/
│       └── documents/
│           ├── extractor.go
│           ├── processor.go
│           └── llm.go
│
├── frontend/
│   └── src/
│       ├── components/
│       ├── services/
│       ├── types/
│       └── App.tsx
│
└── README.md
```

## 🚧 Development Status

FlowPilot is actively under development.

The initial milestone focuses on document upload, AI-powered analysis, structured extraction, and contextual summaries. RAG-based retrieval, source-grounded citations, and additional document intelligence capabilities will be developed incrementally.

## 🎯 Vision

Make document analysis less like reading a 50-page report and more like having a smart, slightly quirky colleague explain what actually matters.

**Less reading. More understanding. A little more personality.**

---

*FlowPilot — Your documents, decoded by AI.*