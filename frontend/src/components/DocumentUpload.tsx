
import { useRef } from 'react';
import { FileText, UploadCloud, X } from 'lucide-react';

interface Props {
  file: File | null;
  onFileChange: (file: File | null) => void;
}

export default function DocumentUpload({ file, onFileChange }: Props) {
  const inputRef = useRef<HTMLInputElement>(null);

  const handleChange = (selected?: File) => {
    if (!selected) return;

    const extension = selected.name.split('.').pop()?.toLowerCase();

    if (!['pdf', 'docx', 'txt'].includes(extension ?? '')) {
      alert('Please select a PDF, DOCX or TXT document.');
      return;
    }

    if (selected.size > 10 * 1024 * 1024) {
      alert('Maximum file size is 10 MB.');
      return;
    }

    onFileChange(selected);
  };

  return (
    <section className="content-card">
      <div className="section-heading">
        <div>
          <h2>Upload document</h2>
          <p>Select a document to extract and summarize.</p>
        </div>
      </div>

      <input
        ref={inputRef}
        type="file"
        accept=".pdf,.docx,.txt"
        hidden
        onChange={(event) => handleChange(event.target.files?.[0])}
      />

      {!file ? (
        <button
          type="button"
          className="upload-area"
          onClick={() => inputRef.current?.click()}
        >
          <UploadCloud size={30} />
          <strong>Click to upload your document</strong>
          <span>PDF, DOCX or TXT · Maximum 10 MB</span>
        </button>
      ) : (
        <div className="file-preview">
          <FileText size={24} />
          <div>
            <strong>{file.name}</strong>
            <span>{(file.size / 1024 / 1024).toFixed(2)} MB</span>
          </div>

          <button
            type="button"
            className="icon-button"
            aria-label="Remove document"
            onClick={() => {
              onFileChange(null);
              if (inputRef.current) inputRef.current.value = '';
            }}
          >
            <X size={17} />
          </button>
        </div>
      )}
    </section>
  );
}
