import {
  useRef,
  useState,
  type DragEvent,
  type FormEvent,
} from 'react';

import {
  askDocument,
  uploadDocument,
} from '../services/api/documentService';

import { ApiError } from '../services/api/apiClient';

import type {
  Citation,
  IngestResult,
} from '../services/api/types';

import {
  ArrowUpRight,
  FileText,
  Sparkles,
  Workflow,
  X,
  MessageCircle,
  Send,
  BookOpen,
} from 'lucide-react';

import './App.css';

const MAX_FILE_SIZE = 10 * 1024 * 1024;
const SUPPORTED_TYPES = ['.pdf', '.docx', '.txt'];

interface ConversationItem {
  question: string;
  answer: string;
  citations: Citation[];
}

export default function App() {
  const inputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState('');
  const [dragging, setDragging] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [ingestResult, setIngestResult] = useState<IngestResult | null>(null);
  const [question, setQuestion] = useState('');
  const [asking, setAsking] = useState(false);
  const [askError, setAskError] = useState('');
  const [messages, setMessages] = useState<ConversationItem[]>([]);
  const selectFile = (selectedFile?: File) => {
    if (!selectedFile) return;

    const extension = `.${selectedFile.name.split('.').pop()?.toLowerCase()}`;

    if (!SUPPORTED_TYPES.includes(extension)) {
      setError('That file type is not supported. Try PDF, DOCX or TXT.');
      return;
    }

    if (selectedFile.size > MAX_FILE_SIZE) {
      setError('Your document must be smaller than 10 MB.');
      return;
    }

    setError('');
    setIngestResult(null);
    setMessages([]);
    setQuestion('');
    setAskError('');
    setFile(selectedFile);

  };

  const handleUpload = async () => {
    if (!file || uploading) return;

    setUploading(true);
    setError('');
    setIngestResult(null);

    try {
      const result = await uploadDocument(file);

      setIngestResult(result);
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'NETWORK' || err.code === 'TIMEOUT') {
          setError(err.message);
        } else {
          switch (err.status) {
            case 400:
              setError('This document could not be accepted. Please check the file and try again.');
              break;

            case 413:
              setError('Your document is too large. Please upload a file smaller than 10 MB.');
              break;

            case 422:
              setError('We could not extract readable content from this document. Try another file.');
              break;

            case 503:
              setError('FlowPilot is temporarily unavailable. Please try again shortly.');
              break;

            default:
              setError('We could not process your document. Please try again.');
          }
        }
      } else {
        setError('Something unexpected happened. Please try again.');
      }
    } finally {
      setUploading(false);
    }
  };

  const handleAsk = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!ingestResult || !question.trim() || asking) return;

    const currentQuestion = question.trim();

    setAsking(true);
    setAskError('');

    try {
      const response = await askDocument(
        ingestResult.document_id,
        {
          question: currentQuestion,
          limit: 5,
        }
      );

      setMessages((previous) => [
        ...previous,
        {
          question: currentQuestion,
          answer: response.answer,
          citations: response.citations,
        },
      ]);

      setQuestion('');
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.code === 'NETWORK' || err.code === 'TIMEOUT') {
          setAskError(err.message);
        } else {
          switch (err.status) {
            case 400:
              setAskError('Please check your question and try again.');
              break;
            case 503:
              setAskError('The AI service is temporarily unavailable.');
              break;
            default:
              setAskError('We could not generate an answer. Please try again.');
          }
        }
      } else {
        setAskError(
          err instanceof Error
            ? err.message
            : 'Something unexpected happened.'
        );
      }
    } finally {
      setAsking(false);
    }
  };

  const handleDrop = (event: DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragging(false);
    selectFile(event.dataTransfer.files[0]);
  };

  const clearFile = () => {
    setFile(null);
    setError('');
    setIngestResult(null);

    setMessages([]);
    setQuestion('');
    setAskError('');

    if (inputRef.current) {
      inputRef.current.value = '';
    }
  };

  return (
    <main className="page-shell">
      <div className="ambient ambient-purple" />
      <div className="ambient ambient-coral" />

      <header className="site-header">
        <a className="brand" href="/" aria-label="FlowPilot home">
          <span className="brand-mark">
            <Workflow size={19} />
          </span>
          <span>flowpilot<span className="brand-accent">.</span></span>
        </a>

        <div className="header-status">
          <span className="status-dot" />
          Ready to explore
        </div>
      </header>

      <section className="hero">
        <div className="hero-copy">
          <div className="eyebrow">
            <Sparkles size={13} />
            DOCUMENT INTELLIGENCE, REIMAGINED
          </div>

          <h1>
            Your documents
            <br />
            have a lot to say.
            <br />
            <span>Let's find out what.</span>
          </h1>

          <p>
            Drop in a document. We'll turn all that information
            into something you can actually use.
          </p>
        </div>

        <div className="hero-art" aria-hidden="true">
          <div className="art-orbit art-orbit-one" />
          <div className="art-orbit art-orbit-two" />
          <div className="art-orbit art-orbit-three" />

          <div className="art-document">
            <div className="art-document-icon">
              <FileText size={27} />
            </div>
            <div className="art-line art-line-long" />
            <div className="art-line art-line-medium" />
            <div className="art-line art-line-short" />
            <div className="art-highlight" />
          </div>

          <div className="art-sparkle sparkle-one">
            <Sparkles size={19} />
          </div>
          <div className="art-sparkle sparkle-two">
            <Sparkles size={13} />
          </div>
          <div className="art-dot" />
        </div>
      </section>

      <section className="upload-section">
        <div
          className={`upload-zone ${dragging ? 'upload-zone-active' : ''} ${
            file ? 'upload-zone-selected' : ''
          }`}
          onDragOver={(event) => {
            event.preventDefault();
            setDragging(true);
          }}
          onDragLeave={() => setDragging(false)}
          onDrop={handleDrop}
        >
          {!file ? (
            <>
              <div className="upload-symbol">
                <FileText size={25} />
                <span>+</span>
              </div>

              <h2>Give your document a new perspective.</h2>

              <p>
                Drag and drop it here, or{' '}
                <button
                  className="browse-button"
                  onClick={() => inputRef.current?.click()}
                >
                  browse files
                </button>
              </p>

              <span className="file-formats">
                PDF · DOCX · TXT <span>—</span> MAX 10 MB
              </span>
            </>
          ) : (
            <div className="selected-file">
              <div className="selected-file-icon">
                <FileText size={24} />
              </div>

              <div className="selected-file-info">
                <strong>{file.name}</strong>
                <span>
                  {ingestResult
                    ? 'Successfully processed'
                    : uploading
                      ? 'Uploading and processing...'
                      : `${(file.size / 1024 / 1024).toFixed(2)} MB · Ready to process`}
                </span>
              </div>

              <button
                className="remove-file"
                onClick={clearFile}
                aria-label="Remove selected file"
                disabled={uploading}
              >
                <X size={17} />
              </button>
            </div>
          )}
        </div>

        <input
          ref={inputRef}
          type="file"
          accept=".pdf,.docx,.txt"
          hidden
          onChange={(event) => selectFile(event.target.files?.[0])}
        />

        {error && (
          <p className="upload-error" role="alert">
            {error}
          </p>
        )}

        <div className="upload-action">
          <button
            className="primary-button"
            disabled={!file || uploading || !!ingestResult}
            onClick={handleUpload}
          >
            {uploading
              ? 'Making sense of your document...'
              : ingestResult
                ? 'Document processed'
                : error
                  ? 'Try again'
                  : "Let's make sense of it"}

            <ArrowUpRight size={17} />
          </button>

          <span>AI-powered · Source-grounded · Yours to explore</span>
        </div>
        {ingestResult && (
          <section className="qa-section">
            <div className="qa-heading">
              <div>
                <div className="eyebrow">
                  <MessageCircle size={14} />
                  DOCUMENT CONVERSATION
                </div>

                <h2>Now, let's get curious.</h2>

                <p>
                  Ask questions and explore the information in your document.
                </p>
              </div>

              <span className="qa-document-badge">
                <FileText size={14} />
                {ingestResult.file_name}
              </span>
            </div>

            {messages.length > 0 && (
              <div className="conversation">
                {messages.map((message, index) => (
                  <article className="conversation-item" key={index}>
                    <div className="user-question">
                      <div className="message-avatar">
                        <MessageCircle size={16} />
                      </div>

                      <p>{message.question}</p>
                    </div>

                    <div className="ai-answer">
                      <div className="answer-heading">
                        <Sparkles size={17} />
                        FlowPilot's answer
                      </div>

                      <p className="answer-text">{message.answer}</p>

                      {message.citations.length > 0 && (
                        <div className="citations">
                          <div className="citations-heading">
                            <BookOpen size={15} />
                            Supporting sources
                          </div>

                          {message.citations.map((citation, citationIndex) => (
                            <div
                              className="citation-card"
                              key={`${index}-${citationIndex}`}
                            >
                              <div className="citation-meta">
                                <FileText size={14} />
                                <strong>{citation.source}</strong>

                                {citation.page !== undefined && (
                                  <span className="citation-page">
                                    Page {citation.page}
                                  </span>
                                )}

                                {citation.field && (
                                  <span className="citation-page">
                                    {citation.field}
                                  </span>
                                )}
                              </div>

                              {citation.quote && (
                                <blockquote>
                                  "{citation.quote}"
                                </blockquote>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  </article>
                ))}
              </div>
            )}

            {asking && (
              <div className="answer-loading" role="status">
                <Sparkles size={17} />
                Looking through your document...
              </div>
            )}

            {askError && (
              <div className="qa-error" role="alert">
                {askError}
              </div>
            )}

            <form className="question-form" onSubmit={handleAsk}>
              <textarea
                value={question}
                onChange={(event) => setQuestion(event.target.value)}
                placeholder="Ask something about your document..."
                rows={2}
                maxLength={4000}
                disabled={asking}
                aria-label="Ask your document a question"
              />

              <div className="question-form-footer">
                <span>Answers are grounded in your uploaded document.</span>

                <button
                  className="ask-button"
                  type="submit"
                  disabled={!question.trim() || asking}
                >
                  {asking ? 'Thinking...' : 'Ask FlowPilot'}
                  <Send size={15} />
                </button>
              </div>
            </form>
          </section>
        )}
        {ingestResult && (
          <div className="upload-success" role="status">
            <div className="success-heading">
              <Sparkles size={18} />
              Your document is ready!
            </div>

            <p>
              Successfully processed {ingestResult.file_name}.
            </p>

            <div className="document-stats">
              <span>{ingestResult.word_count.toLocaleString()} words</span>
              <span>{ingestResult.chunk_count} chunks</span>
              <span>
                {ingestResult.character_count.toLocaleString()} characters
              </span>
            </div>
          </div>
        )}
        
      </section>

      <section className="process-section">
        <div className="process-step">
          <span className="step-number step-coral">01</span>
          <span>Upload</span>
        </div>

        <div className="process-connector" />

        <div className="process-step">
          <span className="step-number step-purple">02</span>
          <span>Understand</span>
        </div>

        <div className="process-connector" />

        <div className="process-step">
          <span className="step-number step-lime">03</span>
          <span>Ask anything</span>
        </div>
      </section>

      <footer className="site-footer">
        <span>Made for curious minds.</span>
        <span>FlowPilot © 2026</span>
      </footer>
    </main>
  );
}