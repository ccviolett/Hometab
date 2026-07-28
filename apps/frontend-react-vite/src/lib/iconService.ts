import { apiClient } from './api-client';

export interface IconCheckResult {
  host: string;
  status: 'ready' | 'failed' | 'unchanged' | 'conflict';
  current_icon_url?: string;
  pending_icon_url?: string;
  error?: string;
}

export interface DomainIconItem {
  host: string;
  icon_path: string;
  content_type: string;
  source: 'auto' | 'manual' | 'user_confirmed' | 'fallback';
  status: 'ready' | 'failed' | 'conflict';
  pending_path?: string;
  last_checked_at?: string;
  error_message?: string;
}

export interface IconRefreshAllResult {
  total_links: number;
  total_hosts: number;
  ready: number;
  unchanged: number;
  failed: number;
  conflicts: number;
  errors?: string[];
}

export class IconService {
  static getIconUrl(url: string, version = 0): string {
    const base = `/api/link-icons/resolve?url=${encodeURIComponent(url)}`;
    return version > 0 ? `${base}&v=${version}` : base;
  }

  static clearCache() {
    // Icons are now cached by the backend, not in browser localStorage.
  }

  static clearExpired() {
    // Icons are now cached by the backend, not in browser localStorage.
  }

  static async checkIcon(url: string): Promise<IconCheckResult> {
    return apiClient.post<IconCheckResult>('api/link-icons/check', { url });
  }

  static async chooseIcon(host: string, choice: 'current' | 'new') {
    return apiClient.post(`api/link-icons/${encodeURIComponent(host)}/choose`, { choice });
  }

  static async refreshAll(): Promise<IconRefreshAllResult> {
    return apiClient.post<IconRefreshAllResult>('api/link-icons/refresh-all');
  }

  static async list(): Promise<DomainIconItem[]> {
    return apiClient.get<DomainIconItem[]>('api/link-icons');
  }

  static async upload(host: string, file: File): Promise<DomainIconItem> {
    const form = new FormData();
    form.append('file', file);
    const response = await fetch(`/api/link-icons/${encodeURIComponent(host)}/upload`, { method: 'POST', body: form });
    if (!response.ok) throw new Error(await response.text());
    return response.json() as Promise<DomainIconItem>;
  }

  static async remove(host: string): Promise<void> {
    await apiClient.delete(`api/link-icons/${encodeURIComponent(host)}`);
  }

  static async retry(host: string): Promise<IconCheckResult> {
    return apiClient.post<IconCheckResult>(`api/link-icons/${encodeURIComponent(host)}/retry`, { url: `https://${host}` });
  }
}

export default IconService;
