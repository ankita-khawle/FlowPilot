
import type { SummaryLength, SummaryResponse, SummaryTone } from '../App';

interface SummarizeRequest {
  file: File;
  tone: SummaryTone;
  length: SummaryLength;
}

export async function summarizeDocument({
  file,
  tone,
  length,
}: SummarizeRequest): Promise<SummaryResponse> {
  const formData = new FormData();

  formData.append('file', file);
  formData.append('tone', tone);
  formData.append('length', length);

  const response = await fetch('/api/documents/summarize', {
    method: 'POST',
    body: formData,
  });

  if (!response.ok) {
    const body = await response.json().catch(() => null);

    throw new Error(
      body?.error || 'Failed to generate document summary.'
    );
  }

  return response.json();
}
