
import { useRef, useState } from 'react';
import {
  FileCheck2,
  FileText,
  LoaderCircle,
  Sparkles,
  UploadCloud,
  X,
} from 'lucide-react';

type DocumentResult = {
  summary: string;
  extracted_data?: Record<string, unknown>;
};

const MAX_FILE_SIZE = 10 * 1024 * 1024;
const ACCEPTED_TYPES = ['.pdf', '.docx', '.txt'];

export default function DocumentSummary() {
  const inputRef = useRef<HTMLInputElement>(null);

  const [file, setFile] = useState<File | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<DocumentResult | null>(null);

  const handleFileChange = (selectedFile?: File) => {
    if (!selectedFile) return;

    const extension = selectedFile.name
      .slice(selectedFile.name.lastIndexOf('.'))
      .toLowerCase();

    if (!ACCEPTED_TYPES.includes(extension)) {
      setError('Please upload a PDF, DOCX, or TXT file.');
      return;
    }

    if (selectedFile.size > MAX_FILE_SIZE) {
      setError('File size must not exceed 10 MB.');
      return;
    }

    setFile(selectedFile);
    setError('');
    setResult(null);
  };

  const handleUpload = async () => {
    if (!file) return;

    setLoading(true);
    setError('');
    setResult(null);

    try {
      const formData = new FormData();
      formData.append('file', file);

      const response = await fetch('/api/documents/summarize', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const body = await response.json().catch(() => null);
        throw new Error(
          body?.error || 'Unable to process the document. Please try again.'
        );
      }

      const data: DocumentResult = await response.json();

      if (!data.summary) {
        throw new Error('The API returned an empty summary.');
      }

      setResult(data);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Something went wrong.'
      );
    } finally {
      setLoading(false);
    }
  };

  const formatValue = (value: unknown): string => {
    if (typeof value === 'string') return value;
    if (value == null) return '—';

    if (typeof value === 'object') {
      return JSON.stringify(value, null, 2);
    }

    return String(value);
  };

  return (
    <section className="document-panel" id="document-intelligence">
      <div className="panel-heading">
        <div>
          <h2>Document intelligence</h2>
          <p>Extract information and generate AI-powered summaries.</p>
        </div>
        <span className="ai-label">
          <Sparkles size={14} /> AI powered
        </span>
      </div>

      <div className="document-upload-area">
        <input
          ref={inputRef}
          type="file"
          accept=".pdf,.docx,.txt"
          hidden
          onChange={(event) => handleFileChange(event.target.files?.[0])}
        />

        {!file ? (
          <div className="upload-empty">
            <div className="upload-icon">
              <UploadCloud size={25} />
            </div>

            <h3>Upload a document</h3>
            <p>
              Select a document to extract its content and generate a
              concise summary.
            </p>

            <button
              className="secondary-button"
              onClick={() => inputRef.current?.click()}
              type="button"
            >
              <FileText size={16} /> Choose document
            </button>

            <span className="upload-hint">
              PDF, DOCX or TXT · Maximum file size 10 MB
            </span>
          </div>
        ) : (
          <div className="selected-document">
            <div className="selected-file">
              <FileText size={22} />
              <div>
                <strong>{file.name}</strong>
                <span>{(file.size / (1024 * 1024)).toFixed(2)} MB</span>
              </div>
              <button
                className="icon-button"
                type="button"
                aria-label="Remove document"
                disabled={loading}
                onClick={() => {
                  setFile(null);
                  setResult(null);
                  setError('');
                  if (inputRef.current) inputRef.current.value = '';
                }}
              >
                <X size={17} />
              </button>
            </div>

            <button
              className="primary-button"
              type="button"
              onClick={handleUpload}
              disabled={loading}
            >
              {loading ? (
                <>
                  <LoaderCircle size={17} className="spin" />
                  Processing document...
                </>
              ) : (
                <>
                  <Sparkles size={17} />
                  Extract and summarize
                </>
              )}
            </button>
          </div>
        )}
      </div>

      {error && (
        <div className="document-error" role="alert">
          {error}
        </div>
      )}

      {result && (
        <div className="document-result" aria-live="polite">
          <div className="result-heading">
            <div>
              <FileCheck2 size={19} />
              <h3>Generated summary</h3>
            </div>
            <span className="result-status">Completed</span>
          </div>

          <p className="summary-text">{result.summary}</p>

          {result.extracted_data &&
            Object.keys(result.extracted_data).length > 0 && (
              <div className="extracted-data">
                <h4>Extracted information</h4>

                <div className="extracted-data-grid">
                  {Object.entries(result.extracted_data).map(
                    ([key, value]) => (
                      <div className="extracted-item" key={key}>
                        <span>
                          {key.replace(/_/g, ' ')}
                        </span>
                        <strong>{formatValue(value)}</strong>
                      </div>
                    )
                  )}
                </div>
              </div>
            )}
        </div>
      )}
    </section>
  );
}
