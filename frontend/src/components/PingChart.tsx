import { useMemo, useState } from 'react';
import { Segmented } from 'antd';
import {
  Chart as ChartJS,
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Filler,
  Tooltip,
  TimeScale,
  Legend,
} from 'chart.js';
import 'chartjs-adapter-dayjs-4';
import { Chart } from 'react-chartjs-2';
import { useChartData } from '../hooks/useHeartbeats';
import type { Heartbeat, ChartPoint } from '../hooks/useHeartbeats';
import { useMonitor } from '../hooks/useMonitors';
import { STATUS_COLORS, STATUS } from '../lib/constants';

ChartJS.register(
  BarController,
  BarElement,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Filler,
  Tooltip,
  TimeScale,
  Legend,
);

interface PingChartProps {
  monitorId: string;
  heartbeats?: Heartbeat[];
}

const PERIODS = [
  { label: 'Recent', value: 0 },
  { label: '3h', value: 3 },
  { label: '6h', value: 6 },
  { label: '24h', value: 24 },
  { label: '7d', value: 168 },
];

function getBarColor(point: { down?: number; up?: number; maintenance?: number }): string {
  if (point.maintenance) {
    return 'rgba(23,71,245,0.41)';
  } else if (!point.down) {
    return '#000';
  } else if (!point.up) {
    return 'rgba(220, 53, 69, 0.41)';
  }
  return 'rgba(245, 182, 23, 0.41)';
}

function getHeartbeatBarColor(status: number): string {
  switch (status) {
    case STATUS.MAINTENANCE:
      return 'rgba(23, 71, 245, 0.41)';
    case STATUS.PENDING:
      return 'rgba(245, 182, 23, 0.41)';
    default:
      return 'rgba(220, 53, 69, 0.41)';
  }
}

interface XYPoint {
  x: string;
  y: number | null;
}

function buildRecentData(heartbeats: Heartbeat[], monitorInterval: number | undefined) {
  const pingData: XYPoint[] = [];
  const downData: XYPoint[] = [];
  const colorData: string[] = [];

  let lastTime: Date | null = null;

  for (const beat of heartbeats) {
    const beatTime = new Date(beat.time);
    const x = beat.time;

    if (lastTime && monitorInterval) {
      const diff = Math.abs(beatTime.getTime() - lastTime.getTime());
      if (diff > monitorInterval * 1000 * 10) {
        const gapStart = new Date(lastTime.getTime() + monitorInterval * 1000).toISOString();
        const gapEnd = new Date(beatTime.getTime() - monitorInterval * 1000).toISOString();
        for (const gx of [gapStart, gapEnd]) {
          pingData.push({ x: gx, y: null });
          downData.push({ x: gx, y: null });
          colorData.push('#000');
        }
      }
    }

    pingData.push({
      x,
      y: beat.status === STATUS.UP && beat.ping ? beat.ping : null,
    });
    downData.push({
      x,
      y: beat.status === STATUS.DOWN || beat.status === STATUS.MAINTENANCE || beat.status === STATUS.PENDING ? 1 : 0,
    });
    colorData.push(getHeartbeatBarColor(beat.status));
    lastTime = beatTime;
  }

  return { pingData, downData, colorData };
}

interface AggPoint {
  timestamp: number;
  up: number;
  down: number;
  avgPing: number;
  minPing: number;
  maxPing: number;
  maintenance?: number;
}

function getAverage(points: AggPoint[]): AggPoint {
  const totalUp = points.reduce((sum, p) => sum + p.up, 0);
  const totalDown = points.reduce((sum, p) => sum + p.down, 0);
  const totalMaintenance = points.reduce((sum, p) => sum + (p.maintenance || 0), 0);
  const totalPing = points.reduce((sum, p) => sum + p.avgPing * p.up, 0);
  const minPing = points.reduce((min, p) => Math.min(min, p.minPing), Infinity);
  const maxPing = points.reduce((max, p) => Math.max(max, p.maxPing), 0);
  const mid = Math.floor(points.length / 2);
  return {
    timestamp: points[mid].timestamp,
    up: totalUp,
    down: totalDown,
    maintenance: totalMaintenance > 0 ? totalMaintenance : undefined,
    avgPing: totalUp > 0 ? totalPing / totalUp : 0,
    minPing,
    maxPing,
  };
}

function buildStatsData(chartPoints: ChartPoint[], monitorInterval: number | undefined, period: number) {
  const avgPingData: XYPoint[] = [];
  const minPingData: XYPoint[] = [];
  const maxPingData: XYPoint[] = [];
  const downData: XYPoint[] = [];
  const colorData: string[] = [];

  const aggregatePoints = period > 6 ? 12 : 4;
  let aggregateBuffer: AggPoint[] = [];
  let lastTime: Date | null = null;

  function pushPoint(dp: AggPoint) {
    const x = new Date(dp.timestamp * 1000).toISOString();
    avgPingData.push({ x, y: dp.up > 0 && dp.avgPing != null ? dp.avgPing : null });
    minPingData.push({ x, y: dp.up > 0 && dp.avgPing != null ? dp.minPing : null });
    maxPingData.push({ x, y: dp.up > 0 && dp.avgPing != null ? dp.maxPing : null });
    downData.push({ x, y: dp.down + (dp.maintenance || 0) });
    colorData.push(getBarColor(dp));
  }

  function pushNulls(x: string) {
    avgPingData.push({ x, y: null });
    minPingData.push({ x, y: null });
    maxPingData.push({ x, y: null });
    downData.push({ x, y: null });
    colorData.push('#000');
  }

  for (const raw of chartPoints) {
    const dp: AggPoint = {
      timestamp: raw.timestamp ?? 0,
      up: raw.up ?? 0,
      down: raw.down ?? 0,
      avgPing: raw.ping ?? 0,
      minPing: raw.pingMin ?? 0,
      maxPing: raw.pingMax ?? 0,
    };

    if (dp.up === 0 && dp.down === 0) continue;

    const beatTime = new Date(dp.timestamp * 1000);

    if (lastTime && monitorInterval) {
      const diff = Math.abs(beatTime.getTime() - lastTime.getTime());
      const oneMinute = 60_000;
      const oneHour = 3_600_000;
      const threshold = period <= 24
        ? Math.max(oneMinute * 10, monitorInterval * 1000 * 10)
        : Math.max(oneHour * 10, monitorInterval * 1000 * 10);

      if (diff > threshold) {
        if (aggregateBuffer.length > 0) {
          pushPoint(getAverage(aggregateBuffer));
          aggregateBuffer = [];
        }
        const gapStart = new Date(lastTime.getTime() + monitorInterval * 1000).toISOString();
        const gapEnd = new Date(dp.timestamp * 1000 + 60_000).toISOString();
        pushNulls(gapStart);
        pushNulls(gapEnd);
      }
    }

    if (dp.up > 0 && chartPoints.length > aggregatePoints * 2) {
      aggregateBuffer.push(dp);
      if (aggregateBuffer.length === aggregatePoints) {
        pushPoint(getAverage(aggregateBuffer));
        aggregateBuffer = aggregateBuffer.slice(Math.floor(aggregatePoints / 2));
      }
    } else {
      if (aggregateBuffer.length > 0) {
        pushPoint(getAverage(aggregateBuffer));
        aggregateBuffer = [];
      }
      pushPoint(dp);
    }

    lastTime = beatTime;
  }

  if (aggregateBuffer.length > 0) {
    pushPoint(getAverage(aggregateBuffer));
  }

  return { avgPingData, minPingData, maxPingData, downData, colorData };
}

export function PingChart({ monitorId, heartbeats = [] }: PingChartProps) {
  const [period, setPeriod] = useState(0);
  const { data: chartPoints } = useChartData(monitorId, period || undefined);
  const { data: monitor } = useMonitor(monitorId);
  const monitorInterval = monitor?.interval;

  const chartData = useMemo(() => {
    if (period === 0) {
      const { pingData, downData, colorData } = buildRecentData(heartbeats, monitorInterval);
      return {
        datasets: [
          {
            label: 'Ping',
            data: pingData,
            fill: 'origin',
            tension: 0.2,
            borderColor: STATUS_COLORS[STATUS.UP],
            backgroundColor: `${STATUS_COLORS[STATUS.UP]}38`,
            yAxisID: 'y',
            pointRadius: 0,
            pointHitRadius: 100,
          },
          {
            type: 'bar' as const,
            label: 'status',
            data: downData,
            borderColor: '#00000000',
            backgroundColor: colorData,
            yAxisID: 'y1',
            barThickness: 'flex' as const,
            barPercentage: 1,
            categoryPercentage: 1,
          },
        ],
      };
    }

    if (!chartPoints) return { datasets: [] };

    const { avgPingData, minPingData, maxPingData, downData, colorData } = buildStatsData(chartPoints, monitorInterval, period);

    return {
      datasets: [
        {
          label: 'Min',
          data: minPingData,
          fill: 'origin',
          tension: 0.2,
          borderColor: '#126331',
          backgroundColor: '#2F9C5914',
          yAxisID: 'y',
          pointRadius: 0,
          pointHitRadius: 100,
        },
        {
          label: 'Avg Ping',
          data: avgPingData,
          fill: 'origin',
          tension: 0.2,
          borderColor: STATUS_COLORS[STATUS.UP],
          backgroundColor: `${STATUS_COLORS[STATUS.UP]}06`,
          yAxisID: 'y',
          pointRadius: 0,
          pointHitRadius: 100,
        },
        {
          label: 'Max',
          data: maxPingData,
          fill: 'origin',
          tension: 0.2,
          borderColor: '#21b55a',
          backgroundColor: '#1E7A4214',
          yAxisID: 'y',
          pointRadius: 0,
          pointHitRadius: 100,
        },
        {
          type: 'bar' as const,
          label: 'status',
          data: downData,
          borderColor: '#00000000',
          backgroundColor: colorData,
          yAxisID: 'y1',
          barThickness: 'flex' as const,
          barPercentage: 1,
          categoryPercentage: 1,
        },
      ],
    };
  }, [period, heartbeats, chartPoints, monitorInterval]);

  const options = useMemo(
    () => ({
      responsive: true,
      maintainAspectRatio: false,
      interaction: { intersect: false, mode: 'nearest' as const },
      elements: {
        point: { radius: 0, hitRadius: 100 },
      },
      scales: {
        x: {
          type: 'time' as const,
          time: {
            minUnit: 'minute' as const,
            round: 'second' as const,
            tooltipFormat: 'YYYY-MM-DD HH:mm:ss',
            displayFormats: {
              minute: 'HH:mm',
              hour: 'MM-DD HH:mm',
            },
          },
          ticks: {
            sampleSize: 3,
            maxRotation: 0,
            autoSkipPadding: 30,
          },
          grid: { display: false },
        },
        y: {
          beginAtZero: true,
          title: { display: true, text: 'ms' },
        },
        y1: {
          display: false,
          position: 'right' as const,
          grid: { drawOnChartArea: false },
          min: 0,
          max: 1,
        },
      },
      plugins: {
        tooltip: {
          filter: (tooltipItem: { datasetIndex: number; chart: { data: { datasets: { type?: string }[] } } }) => {
            const ds = tooltipItem.chart.data.datasets[tooltipItem.datasetIndex];
            return ds && ds.type !== 'bar';
          },
          callbacks: {
            label: (ctx: { dataset: { label?: string }; parsed: { y: number | null } }) =>
              `${ctx.dataset.label}: ${ctx.parsed.y != null ? `${Math.round(ctx.parsed.y)} ms` : 'N/A'}`,
          },
        },
        legend: {
          display: true,
          position: 'top' as const,
          align: 'start' as const,
          labels: {
            filter: (legendItem: { datasetIndex?: number }, data: { datasets: { type?: string }[] }) => {
              if (legendItem.datasetIndex == null) return true;
              const ds = data.datasets[legendItem.datasetIndex];
              return ds && ds.type !== 'bar';
            },
          },
        },
      },
    }),
    [],
  );

  return (
    <div>
      <div style={{ marginBottom: 12 }}>
        <Segmented
          options={PERIODS.map(p => ({ label: p.label, value: p.value }))}
          value={period}
          onChange={v => setPeriod(v as number)}
          size="small"
        />
      </div>
      <div style={{ height: 250 }}>
        <Chart type="line" data={chartData} options={options} />
      </div>
    </div>
  );
}
