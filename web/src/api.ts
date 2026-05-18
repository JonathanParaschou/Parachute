export type StorageMetadata = {
  platform: string;
  total_storage: number;
  used_storage: number;
  free_storage: number;
  available_storage: number;
  drives: Drive[];
};

export type Drive = {
  name: string;
  model: string;
  serial: string;
  size_bytes: number;
  bus: string;
  is_ssd: boolean | null;
};

export type StorageRoot = {
  id: string;
  path: string;
  limit_bytes: number;
  enabled: boolean;
  created_at: string;
};

export type UploadRecord = {
  id: string;
  original_name: string;
  stored_name: string;
  root_id: string;
  root_path: string;
  object_path: string;
  metadata_path: string;
  size_bytes: number;
  content_type: string;
  uploaded_at: string;
};

export type DashboardData = {
  metadata: StorageMetadata;
  roots: StorageRoot[];
  uploads: UploadRecord[];
  remote: RemoteAccessStatus;
};

export type RemoteAccessStatus = {
  local_url: string;
  lan_urls: string[];
  tailscale: RemoteOption;
  recommended: string;
  server_port: number;
  machine_name: string;
  remote_warnings: string[];
};

export type RemoteOption = {
  available: boolean;
  configured: boolean;
  url?: string;
  address?: string;
  message?: string;
};

async function loadJSON<T>(url: string): Promise<T> {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(await response.text());
  }
  return response.json() as Promise<T>;
}

export async function loadDashboard(): Promise<DashboardData> {
  const [metadata, roots, uploads, remote] = await Promise.all([
    loadJSON<StorageMetadata>("/storage-metadata"),
    loadJSON<StorageRoot[]>("/storage-roots"),
    loadJSON<UploadRecord[]>("/uploads"),
    loadJSON<RemoteAccessStatus>("/remote-access")
  ]);

  return {
    metadata,
    roots: roots ?? [],
    uploads: uploads ?? [],
    remote
  };
}

export async function uploadFile(file: File): Promise<UploadRecord> {
  const formData = new FormData();
  formData.append("uploaded_file", file);

  const response = await fetch("/upload", {
    method: "POST",
    body: formData
  });

  if (!response.ok) {
    throw new Error(await response.text());
  }

  return response.json() as Promise<UploadRecord>;
}
