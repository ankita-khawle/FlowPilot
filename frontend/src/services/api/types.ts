
export interface IngestResult {
  document_id: string;
  file_name: string;
  file_type: string;
  status: string;
  word_count: number;
  character_count: number;
  chunk_count: number;
}

export interface AnswerResponse {
  answer: string;
  citations: Citation[];
}

export interface Citation {
  source: string;
  page?: number;
  quote?: string;
  field?: string;
}

export interface ApiErrorResponse {
  error: string;
}

export interface AskDocumentRequest {
  question: string;
  limit?: number;
}
