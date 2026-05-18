import { useCallback, useEffect, useMemo, useState } from "react";
import {
  Activity,
  AlertCircle,
  CheckCircle2,
  Cpu,
  Database,
  Gauge,
  HardDrive,
  Layers3,
  RefreshCw,
  Server,
  UploadCloud
} from "lucide-react";
import {
  ArcElement,
  BarElement,
  CategoryScale,
  Chart as ChartJS,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip
} from "chart.js";
import { Bar, Doughnut, Line } from "react-chartjs-2";
import { DashboardData, Drive, StorageRoot, UploadRecord, loadDashboard, uploadFile } from "./api";
import { formatBytes, formatDate } from "./format";

ChartJS.register(
  ArcElement,
  BarElement,
  CategoryScale,
  Filler,
  Legend,
  LinearScale,
  LineElement,
  PointElement,
  Tooltip
);

type LoadState = "loading" | "ready" | "error";
type Page = "insights" | "rig";

const compactChartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: false
    },
    tooltip: {
      backgroundColor: "#101828",
      padding: 10,
      titleColor: "#ffffff",
      bodyColor: "#ffffff"
    }
  },
  scales: {
    x: {
      grid: {
        display: false
      },
      ticks: {
        color: "#667085"
      }
    },
    y: {
      beginAtZero: true,
      grid: {
        color: "#eef2f6"
      },
      ticks: {
        color: "#667085"
      }
    }
  }
};

export function App() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [page, setPage] = useState<Page>("insights");
  const [loadState, setLoadState] = useState<LoadState>("loading");
  const [error, setError] = useState<string>("");
  const [uploadOpen, setUploadOpen] = useState(false);
  const [uploadStatus, setUploadStatus] = useState<string>("Ready.");
  const [uploadError, setUploadError] = useState(false);
  const [uploading, setUploading] = useState(false);

  const refresh = useCallback(async () => {
    setLoadState((current) => (current === "ready" ? "ready" : "loading"));
    try {
      const nextData = await loadDashboard();
      setData(nextData);
      setError("");
      setLoadState("ready");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Unable to load dashboard data.");
      setLoadState("error");
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const nodeStats = useMemo(() => getNodeStats(data), [data]);

  async function handleUpload(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = event.currentTarget;
    const formData = new FormData(form);
    const file = formData.get("uploaded_file");

    if (!(file instanceof File) || file.size === 0) {
      setUploadStatus("Choose a file before uploading.");
      setUploadError(true);
      return;
    }

    setUploading(true);
    setUploadError(false);
    setUploadStatus(`Uploading ${file.name}...`);

    try {
      const record = await uploadFile(file);
      setUploadStatus(`Saved ${record.original_name} (${formatBytes(record.size_bytes)}).`);
      form.reset();
      await refresh();
    } catch (err) {
      setUploadStatus(err instanceof Error ? err.message : "Upload failed.");
      setUploadError(true);
    } finally {
      setUploading(false);
    }
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">ParaChute Node</p>
          <h1>{page === "insights" ? "Node Insights" : "Rig Stats"}</h1>
          <p className="subtle">
            {loadState === "ready" && data
              ? `Node online on ${data.metadata.platform || "this system"}.`
              : "Checking node status..."}
          </p>
        </div>
        <div className="toolbar">
          <button className="quiet-button" type="button" onClick={() => void refresh()} aria-label="Refresh dashboard">
            <RefreshCw size={17} />
            Refresh
          </button>
          <button className="primary-action" type="button" onClick={() => setUploadOpen((open) => !open)}>
            <UploadCloud size={17} />
            Upload
          </button>
        </div>
      </header>

      <nav className="page-tabs" aria-label="Dashboard pages">
        <button className={page === "insights" ? "active" : ""} type="button" onClick={() => setPage("insights")}>
          <Gauge size={16} />
          Node Insights
        </button>
        <button className={page === "rig" ? "active" : ""} type="button" onClick={() => setPage("rig")}>
          <Server size={16} />
          Rig Stats
        </button>
      </nav>

      {loadState === "error" && (
        <div className="banner error-banner">
          <AlertCircle size={18} />
          <span>{error}</span>
        </div>
      )}

      {uploadOpen && (
        <section className="upload-strip">
          <form onSubmit={(event) => void handleUpload(event)}>
            <label htmlFor="uploaded_file">File</label>
            <input id="uploaded_file" name="uploaded_file" type="file" required />
            <button className="primary-button" type="submit" disabled={uploading}>
              {uploading ? "Uploading..." : "Save to Node"}
            </button>
          </form>
          <div className={uploadError ? "status error-status" : "status"}>{uploadStatus}</div>
        </section>
      )}

      {page === "insights" ? <NodeInsights data={data} stats={nodeStats} /> : <RigStats data={data} />}
    </main>
  );
}

function NodeInsights({ data, stats }: { data: DashboardData | null; stats: NodeStats }) {
  const utilizationData = useMemo(
    () => ({
      labels: ["Used", "Free"],
      datasets: [
        {
          data: [stats.usedBytes, stats.freeBytes],
          backgroundColor: ["#0f766e", "#d7dde8"],
          borderColor: ["#ffffff", "#ffffff"],
          borderWidth: 4,
          hoverOffset: 4
        }
      ]
    }),
    [stats.freeBytes, stats.usedBytes]
  );

  const rootData = useMemo(
    () => ({
      labels: data?.roots.map((root) => root.path.split(/[\\/]/).filter(Boolean).slice(-2).join("/")) ?? [],
      datasets: [
        {
          label: "Allocated",
          data: data?.roots.map((root) => root.limit_bytes) ?? [],
          backgroundColor: "#2563eb",
          borderRadius: 6
        }
      ]
    }),
    [data?.roots]
  );

  const uploadActivity = useMemo(() => buildUploadActivity(data?.uploads ?? []), [data?.uploads]);
  const cumulative = useMemo(() => buildCumulativeSeries(data?.uploads ?? []), [data?.uploads]);

  return (
    <>
      <section className="hero-panel">
        <div>
          <p className="eyebrow">Allocated Capacity</p>
          <strong>{formatBytes(stats.allocatedBytes)}</strong>
          <span>
            {formatBytes(stats.usedBytes)} used, {formatBytes(stats.freeBytes)} available
          </span>
        </div>
        <div className="hero-meter" aria-label={`${stats.usedPercent}% utilization`}>
          <span style={{ width: `${stats.usedPercent}%` }} />
        </div>
      </section>

      <section className="metrics-grid">
        <Metric label="Allocated storage" value={formatBytes(stats.allocatedBytes)} icon={<Database size={19} />} />
        <Metric label="Used / Free" value={`${stats.usedPercent}% used`} detail={formatBytes(stats.freeBytes)} icon={<Gauge size={19} />} />
        <Metric label="Storage roots" value={String(data?.roots.length ?? 0)} icon={<Layers3 size={19} />} />
        <Metric label="Upload count" value={String(data?.uploads.length ?? 0)} icon={<CheckCircle2 size={19} />} />
      </section>

      <section className="insights-grid">
        <article className="panel chart-card compact-chart">
          <PanelTitle title="Utilization" detail={`${stats.usedPercent}% of allocated capacity`} />
          <div className="donut-wrap">
            <Doughnut
              data={utilizationData}
              options={{
                responsive: true,
                maintainAspectRatio: false,
                cutout: "72%",
                plugins: {
                  legend: {
                    position: "bottom",
                    labels: {
                      boxWidth: 10,
                      color: "#344054",
                      usePointStyle: true
                    }
                  },
                  tooltip: {
                    callbacks: {
                      label: (context) => `${context.label}: ${formatBytes(Number(context.raw))}`
                    }
                  }
                }
              }}
            />
            <div className="donut-center">
              <strong>{stats.usedPercent}%</strong>
              <span>used</span>
            </div>
          </div>
        </article>

        <article className="panel chart-card">
          <PanelTitle title="Upload Activity" detail="Daily uploaded volume" />
          <div className="chart-box">
            <Bar
              data={{
                labels: uploadActivity.labels,
                datasets: [
                  {
                    label: "Uploaded",
                    data: uploadActivity.values,
                    backgroundColor: "#0f766e",
                    borderRadius: 6
                  }
                ]
              }}
              options={{
                ...compactChartOptions,
                plugins: {
                  ...compactChartOptions.plugins,
                  tooltip: {
                    callbacks: {
                      label: (context) => formatBytes(Number(context.raw))
                    }
                  }
                }
              }}
            />
          </div>
        </article>

        <article className="panel chart-card">
          <PanelTitle title="Stored Data Trend" detail="Cumulative upload storage" />
          <div className="chart-box">
            <Line
              data={{
                labels: cumulative.labels,
                datasets: [
                  {
                    label: "Stored",
                    data: cumulative.values,
                    borderColor: "#2563eb",
                    backgroundColor: "rgba(37, 99, 235, 0.12)",
                    fill: true,
                    tension: 0.35,
                    pointRadius: 3
                  }
                ]
              }}
              options={{
                ...compactChartOptions,
                plugins: {
                  ...compactChartOptions.plugins,
                  tooltip: {
                    callbacks: {
                      label: (context) => formatBytes(Number(context.raw))
                    }
                  }
                }
              }}
            />
          </div>
        </article>

        <article className="panel chart-card">
          <PanelTitle title="Root Allocation" detail="Configured capacity by root" />
          <div className="chart-box">
            {data?.roots.length ? (
              <Bar
                data={rootData}
                options={{
                  ...compactChartOptions,
                  indexAxis: "y",
                  plugins: {
                    ...compactChartOptions.plugins,
                    tooltip: {
                      callbacks: {
                        label: (context) => formatBytes(Number(context.raw))
                      }
                    }
                  }
                }}
              />
            ) : (
              <EmptyState text="No storage roots configured." />
            )}
          </div>
        </article>
      </section>
    </>
  );
}

function RigStats({ data }: { data: DashboardData | null }) {
  const metadata = data?.metadata;
  const totalDriveBytes = metadata?.drives.reduce((sum, drive) => sum + drive.size_bytes, 0) ?? 0;

  return (
    <section className="rig-layout">
      <article className="panel system-card">
        <PanelTitle title="Machine" detail="Host-level disk and platform metadata" />
        <div className="system-grid">
          <InfoTile label="Platform" value={metadata?.platform || "Unknown"} icon={<Cpu size={18} />} />
          <InfoTile label="System disk" value={formatBytes(metadata?.total_storage)} icon={<HardDrive size={18} />} />
          <InfoTile label="System free" value={formatBytes(metadata?.free_storage)} icon={<Database size={18} />} />
          <InfoTile label="Drive capacity" value={formatBytes(totalDriveBytes)} icon={<Layers3 size={18} />} />
        </div>
      </article>

      <article className="panel">
        <PanelTitle title="Drives" detail={`${metadata?.drives.length ?? 0} devices reported`} />
        <div className="drive-grid">
          {(metadata?.drives ?? []).length === 0 ? (
            <EmptyState text="No drive details available." />
          ) : (
            metadata?.drives.map((drive) => <DriveCard drive={drive} key={`${drive.name}-${drive.serial}`} />)
          )}
        </div>
      </article>
    </section>
  );
}

function Metric({
  label,
  value,
  detail,
  icon
}: {
  label: string;
  value: string;
  detail?: string;
  icon: React.ReactNode;
}) {
  return (
    <article className="metric">
      <div className="metric-icon">{icon}</div>
      <span>{label}</span>
      <strong>{value}</strong>
      {detail && <small>{detail}</small>}
    </article>
  );
}

function PanelTitle({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="panel-heading">
      <div>
        <h2>{title}</h2>
        <p>{detail}</p>
      </div>
    </div>
  );
}

function InfoTile({ label, value, icon }: { label: string; value: string; icon: React.ReactNode }) {
  return (
    <div className="info-tile">
      <div className="info-icon">{icon}</div>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function DriveCard({ drive }: { drive: Drive }) {
  return (
    <div className="drive-card">
      <div>
        <strong>{drive.model || drive.name || "Drive"}</strong>
        <small>{drive.name || "Unknown device"}</small>
      </div>
      <div className="drive-meta">
        <span>{formatBytes(drive.size_bytes)}</span>
        <span>{drive.bus || "Unknown bus"}</span>
        {drive.is_ssd !== null && <span>{drive.is_ssd ? "SSD" : "HDD"}</span>}
      </div>
      {drive.serial && <small className="serial">Serial {drive.serial}</small>}
    </div>
  );
}

function EmptyState({ text }: { text: string }) {
  return <div className="empty-state">{text}</div>;
}

type NodeStats = {
  allocatedBytes: number;
  usedBytes: number;
  freeBytes: number;
  usedPercent: number;
};

function getNodeStats(data: DashboardData | null): NodeStats {
  const allocatedBytes = data?.roots.reduce((sum, root) => sum + root.limit_bytes, 0) ?? 0;
  const usedBytes = data?.uploads.reduce((sum, upload) => sum + upload.size_bytes, 0) ?? 0;
  const freeBytes = Math.max(0, allocatedBytes - usedBytes);
  const usedPercent = allocatedBytes > 0 ? Math.min(100, Math.round((usedBytes / allocatedBytes) * 100)) : 0;

  return {
    allocatedBytes,
    usedBytes,
    freeBytes,
    usedPercent
  };
}

function buildUploadActivity(uploads: UploadRecord[]) {
  const labels = lastDays(7);
  const values = labels.map((label) =>
    uploads
      .filter((upload) => dayLabel(upload.uploaded_at) === label)
      .reduce((sum, upload) => sum + upload.size_bytes, 0)
  );
  return { labels, values };
}

function buildCumulativeSeries(uploads: UploadRecord[]) {
  const labels = lastDays(14);
  let runningTotal = uploads
    .filter((upload) => dayLabel(upload.uploaded_at) < labels[0])
    .reduce((sum, upload) => sum + upload.size_bytes, 0);

  const values = labels.map((label) => {
    runningTotal += uploads
      .filter((upload) => dayLabel(upload.uploaded_at) === label)
      .reduce((sum, upload) => sum + upload.size_bytes, 0);
    return runningTotal;
  });

  return { labels, values };
}

function lastDays(count: number) {
  return Array.from({ length: count }, (_, index) => {
    const date = new Date();
    date.setDate(date.getDate() - (count - index - 1));
    return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
  });
}

function dayLabel(value: string) {
  return new Date(value).toLocaleDateString(undefined, { month: "short", day: "numeric" });
}
