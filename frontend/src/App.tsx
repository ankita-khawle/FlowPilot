
import { useState } from 'react';
import { Sparkles, Workflow } from 'lucide-react';

import DocumentUpload from './components/DocumentUpload';
import SummaryPreferences from './components/SummaryPreferences';
import SummaryResult from './components/SummaryResult';
import { summarizeDocument } from './services/documentService';

export type SummaryTone =
  | 'professional'
  | 'casual'
  | 'technical'
  | 'executive';

export type SummaryLength = 'short' | 'medium' | 'detailed';

export interface SummaryResponse {
  summary: string;
  extracted_data?: Record<string, unknown>;
}

export default function App() {
  const [file, setFile] = useState<File | null>(null);
  const [tone, setTone] = useState<SummaryTone>('professional');
  const [length, setLength] = useState<SummaryLength>('medium');
  const [result, setResult] = useState<SummaryResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleProcess = async () => {
    if (!file || loading) return;

    setLoading(true);
    setError('');
    setResult(null);

    try {
      const response = await summarizeDocument({
        file,
        tone,
        length,
      });

      setResult(response);
    } catch (err) {
      setError(
        err instanceof Error
          ? err.message
          : 'Unable to process document.'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="app-shell">
      <header className="app-header">
        <a className="brand" href="#home">
          <span className="brand-mark">
            <Workflow size={21} />
          </span>
          <span>flowpilot<span className="brand-period">.</span></span>
        </a>

        <span className="header-label">
          <Sparkles size={15} />
          AI DOCUMENT INTELLIGENCE
        </span>
      </header>

      <section className="dashboard" id="home">
        <div className="page-heading">
          <div>
            <span className="eyebrow">DOCUMENT WORKSPACE</span>
            <h1>Document Summarizer</h1>
            <p>
              Extract information and generate summaries
              tailored to your needs.
            </p>
          </div>
        </div>

        <DocumentUpload
          file={file}
          onFileChange={(nextFile) => {
            setFile(nextFile);
            setResult(null);
            setError('');
          }}
        />

        <SummaryPreferences
          tone={tone}
          length={length}
          onToneChange={setTone}
          onLengthChange={setLength}
        />

        <button
          className="primary-button process-button"
          disabled={!file || loading}
          onClick={handleProcess}
        >
          <Sparkles size={17} />
          {loading ? 'Processing document...' : 'Generate summary'}
        </button>

        {error && (
          <div className="document-error" role="alert">
            {error}
          </div>
        )}

        <SummaryResult
          result={result}
          loading={loading}
        />
      </section>
    </main>
  );
}
