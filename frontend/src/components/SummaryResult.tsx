
import { FileCheck2, LoaderCircle, Sparkles } from 'lucide-react';
import type { SummaryResponse } from '../App';

interface Props {
  result: SummaryResponse | null;
  loading: boolean;
}

export default function SummaryResult({ result, loading }: Props) {
  return (
    <section className="content-card summary-section">
      <div className="section-heading">
        <div>
          <h2>Generated summary</h2>
          <p>Extracted information and document insights.</p>
        </div>
        <Sparkles size={19} />
      </div>

      {loading ? (
        <div className="summary-placeholder">
          <LoaderCircle className="spin" size={25} />
          <p>Extracting document content and generating summary...</p>
        </div>
      ) : result ? (
        <>
          <div className="summary-content">
            <div className="result-title">
              <FileCheck2 size={18} />
              Summary
            </div>
            <p>{result.summary}</p>
          </div>

          {result.extracted_data &&
            Object.keys(result.extracted_data).length > 0 && (
              <div className="extracted-content">
                <h3>Extracted information</h3>

                {Object.entries(result.extracted_data).map(([key, value]) => (
                  <div className="extracted-row" key={key}>
                    <strong>{key.replace(/_/g, ' ')}</strong>
                    <span>
                      {typeof value === 'object'
                        ? JSON.stringify(value, null, 2)
                        : String(value ?? '—')}
                    </span>
                  </div>
                ))}
              </div>
            )}
        </>
      ) : (
        <div className="summary-placeholder">
          <FileTextIcon />
          <p>Your generated summary will appear here.</p>
        </div>
      )}
    </section>
  );
}

function FileTextIcon() {
  return <FileCheck2 size={25} />;
}
