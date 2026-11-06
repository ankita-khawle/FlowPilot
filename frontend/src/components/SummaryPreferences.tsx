
import type { SummaryLength, SummaryTone } from '../types/document';

interface Props {
  tone: SummaryTone;
  length: SummaryLength;
  onToneChange: (tone: SummaryTone) => void;
  onLengthChange: (length: SummaryLength) => void;
}

export default function SummaryPreferences({
  tone,
  length,
  onToneChange,
  onLengthChange,
}: Props) {
  return (
    <section className="content-card">
      <div className="section-heading">
        <div>
          <h2>Summary preferences</h2>
          <p>Customize how the AI generates your summary.</p>
        </div>
      </div>

      <div className="preference-group">
        <label htmlFor="summary-tone">Summary tone</label>

        <select
          id="summary-tone"
          value={tone}
          onChange={(event) =>
            onToneChange(event.target.value as SummaryTone)
          }
        >
          <option value="professional">Professional</option>
          <option value="casual">Casual</option>
          <option value="technical">Technical</option>
          <option value="executive">Executive</option>
        </select>
      </div>

      <div className="preference-group">
        <label htmlFor="summary-length">Summary length</label>

        <select
          id="summary-length"
          value={length}
          onChange={(event) =>
            onLengthChange(event.target.value as SummaryLength)
          }
        >
          <option value="short">Short</option>
          <option value="medium">Medium</option>
          <option value="detailed">Detailed</option>
        </select>
      </div>
    </section>
  );
}
