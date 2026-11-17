import { useRef, useState } from 'react';
import {
  ArrowUpRight,
  FileText,
  Sparkles,
  Workflow,
  X,
} from 'lucide-react';

import './App.css';

const MAX_FILE_SIZE = 10 * 1024 * 1024;
const SUPPORTED_TYPES = ['.pdf', '.docx', '.txt'];

export default function App() {
  const inputRef = useRef<HTMLInputElement>(null);
  const [file, setFile] = useState<File | null>(null);
  const [error, setError] = useState('');
  const [dragging, setDragging] = useState(false);

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
    setFile(selectedFile);
  };

  const handleDrop = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setDragging(false);
    selectFile(event.dataTransfer.files[0]);
  };

  const clearFile = () => {
    setFile(null);
    setError('');

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
                <span>{(file.size / 1024 / 1024).toFixed(2)} MB · Ready to process</span>
              </div>

              <button
                className="remove-file"
                onClick={clearFile}
                aria-label="Remove selected file"
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
            disabled={!file}
            onClick={() => {
              // Document ingestion will be connected in the next commit.
            }}
          >
            Let's make sense of it
            <ArrowUpRight size={17} />
          </button>

          <span>AI-powered · Source-grounded · Yours to explore</span>
        </div>
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