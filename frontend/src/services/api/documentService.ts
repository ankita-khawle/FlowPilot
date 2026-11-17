
import { apiRequest, ApiError } from './apiClient';
import type { IngestResult } from './types';

function isIngestResult(value: unknown): value is IngestResult {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  const result = value as Record<string, unknown>;

  return (
    typeof result.document_id === 'string' &&
    result.document_id.length > 0 &&
    typeof result.file_name === 'string' &&
    typeof result.file_type === 'string' &&
    typeof result.status === 'string' &&
    typeof result.word_count === 'number' &&
    typeof result.character_count === 'number' &&
    typeof result.chunk_count === 'number'
  );
}

export async function uploadDocument(
  file: File
): Promise<IngestResult> {
  const formData = new FormData();

  formData.append('file', file);

  const result = await apiRequest<unknown>(
    '/api/rag/documents',
    {
      method: 'POST',
      body: formData,
    }
  );

  if (!isIngestResult(result)) {
    throw new ApiError(
      'The server returned an unexpected document response.',
      200,
      'INVALID_RESPONSE'
    );
  }

  return result;
}
