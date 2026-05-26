import { useEffect, useMemo, useState } from 'react';

// material-ui
import Box from '@mui/material/Box';
import { useColorScheme, useTheme } from '@mui/material/styles';

// third party
import { ApexOptions } from 'apexcharts';
import ReactApexChart from 'react-apexcharts';

// project imports
import { ThemeMode } from 'config';
import useConfig from 'hooks/useConfig';
import axiosServices from 'utils/axios';

interface StatPoint {
  time: string;
  total_requests: number;
  earnings_usd: number;
}

interface ChartStatsResponse {
  granularity: 'hour' | 'day';
  points: StatPoint[];
}

type PeriodMode = 'day' | 'week' | 'month' | 'custom';

interface RequestsAndEarningsProps {
  periodMode?: PeriodMode;
  customFrom?: string;
  customTo?: string;
  projectId?: string | null;
}

const COLOR_REQUESTS = '#2196f3';
const COLOR_EARNINGS = '#4caf50';

function buildChartQuery(mode?: PeriodMode, from?: string, to?: string, projectId?: string | null): string {
  let base: string;
  if (!mode) {
    base = '?days=7';
  } else if (mode === 'custom') {
    if (from && to) base = `?from=${from}&to=${to}`;
    else return '';
  } else {
    base = `?period=${mode}`;
  }
  if (projectId) base += `&project_id=${projectId}`;
  return base;
}

function fillGaps(points: StatPoint[], granularity: 'hour' | 'day'): StatPoint[] {
  if (points.length < 2) return points;

  const pointMap = new Map<string, StatPoint>();
  for (const p of points) {
    const d = new Date(p.time);
    const key =
      granularity === 'hour'
        ? `${d.getUTCFullYear()}-${d.getUTCMonth()}-${d.getUTCDate()}-${d.getUTCHours()}`
        : `${d.getUTCFullYear()}-${d.getUTCMonth()}-${d.getUTCDate()}`;
    pointMap.set(key, p);
  }

  const current = new Date(points[0].time);
  const last = new Date(points[points.length - 1].time);

  if (granularity === 'hour') {
    current.setUTCMinutes(0, 0, 0);
    last.setUTCMinutes(0, 0, 0);
  } else {
    current.setUTCHours(0, 0, 0, 0);
    last.setUTCHours(0, 0, 0, 0);
  }

  const result: StatPoint[] = [];
  while (current <= last) {
    const key =
      granularity === 'hour'
        ? `${current.getUTCFullYear()}-${current.getUTCMonth()}-${current.getUTCDate()}-${current.getUTCHours()}`
        : `${current.getUTCFullYear()}-${current.getUTCMonth()}-${current.getUTCDate()}`;

    result.push(pointMap.get(key) ?? { time: current.toISOString(), total_requests: 0, earnings_usd: 0 });

    if (granularity === 'hour') {
      current.setUTCHours(current.getUTCHours() + 1);
    } else {
      current.setUTCDate(current.getUTCDate() + 1);
    }
  }

  return result;
}

// ==============================|| AREA CHART ||============================== //

export default function RequestsAndEarnings({ periodMode, customFrom, customTo, projectId }: RequestsAndEarningsProps) {
  const theme = useTheme();
  const { colorScheme } = useColorScheme();
  const {
    state: { fontFamily }
  } = useConfig();

  const isDark = colorScheme === ThemeMode.DARK;
  const textPrimary = theme.palette.text.primary;
  const gridColor = theme.palette.divider;

  const [granularity, setGranularity] = useState<'hour' | 'day'>('day');
  const [series, setSeries] = useState([
    { name: 'Requests', data: [] as { x: number; y: number }[] },
    { name: 'Earnings (USD)', data: [] as { x: number; y: number }[] }
  ]);

  useEffect(() => {
    const query = buildChartQuery(periodMode, customFrom, customTo, projectId);
    if (!query) return;
    axiosServices
      .get<ChartStatsResponse>(`/api/v1/stats/daily${query}`)
      .then((res) => {
        const { granularity: g, points } = res.data;
        const filled = fillGaps(points, g);
        setGranularity(g);
        setSeries([
          { name: 'Requests', data: filled.map((p) => ({ x: new Date(p.time).getTime(), y: p.total_requests })) },
          { name: 'Earnings (USD)', data: filled.map((p) => ({ x: new Date(p.time).getTime(), y: parseFloat(p.earnings_usd.toFixed(4)) })) }
        ]);
      })
      .catch(() => {});
  }, [periodMode, customFrom, customTo, projectId]);

  const options = useMemo((): ApexOptions => {
    const isHourly = granularity === 'hour';
    return {
      chart: {
        height: 350,
        type: 'area',
        background: 'transparent',
        fontFamily,
        foreColor: textPrimary,
        toolbar: { show: true },
        zoom: { enabled: false }
      },
      colors: [COLOR_REQUESTS, COLOR_EARNINGS],
      dataLabels: { enabled: false },
      stroke: { curve: 'smooth', width: [2, 2] },
      fill: {
        type: 'gradient',
        gradient: {
          shadeIntensity: 1,
          opacityFrom: 0.35,
          opacityTo: 0.05,
          stops: [0, 100],
          colorStops: [
            [
              { offset: 0, color: COLOR_REQUESTS, opacity: 0.35 },
              { offset: 100, color: COLOR_REQUESTS, opacity: 0.05 }
            ],
            [
              { offset: 0, color: COLOR_EARNINGS, opacity: 0.35 },
              { offset: 100, color: COLOR_EARNINGS, opacity: 0.05 }
            ]
          ]
        }
      },
      markers: {
        size: 4,
        colors: [COLOR_REQUESTS, COLOR_EARNINGS],
        strokeColors: isDark ? '#1a223f' : '#ffffff',
        strokeWidth: 2,
        hover: { size: 7 }
      },
      xaxis: {
        type: 'datetime',
        tickPlacement: 'on',
        axisBorder: { show: true, color: gridColor },
        axisTicks: { show: true, color: gridColor },
        labels: {
          show: true,
          rotate: 0,
          hideOverlappingLabels: true,
          style: { colors: textPrimary },
          datetimeUTC: false,
          format: isHourly ? 'HH:mm' : 'dd MMM',
          datetimeFormatter: isHourly
            ? { year: 'yyyy', month: 'MMM dd', day: 'dd MMM', hour: 'HH:mm', minute: 'HH:mm' }
            : { year: 'yyyy', month: 'MMM yyyy', day: 'dd MMM', hour: 'HH:mm' }
        },
        tooltip: { enabled: true }
      },
      yaxis: [
        {
          labels: {
            style: { colors: textPrimary },
            formatter: (v) => String(Math.round(v))
          },
          title: { text: 'Requests', style: { color: COLOR_REQUESTS } }
        },
        {
          opposite: true,
          labels: {
            style: { colors: textPrimary },
            formatter: (v) => `$${v.toFixed(4)}`
          },
          title: { text: 'Earnings (USD)', style: { color: COLOR_EARNINGS } }
        }
      ],
      tooltip: {
        enabled: true,
        shared: true,
        intersect: false,
        followCursor: true,
        theme: isDark ? 'dark' : 'light',
        x: { format: isHourly ? 'dd MMM HH:mm' : 'dd MMM yyyy' },
        y: [
          { formatter: (v) => (v !== undefined ? String(Math.round(v)) + ' req' : '—') },
          { formatter: (v) => (v !== undefined ? `$${v.toFixed(4)}` : '—') }
        ]
      },
      grid: {
        borderColor: gridColor,
        strokeDashArray: 4
      },
      legend: {
        show: true,
        position: 'top',
        horizontalAlign: 'left',
        offsetX: 0,
        offsetY: 0,
        labels: { colors: textPrimary },
        markers: {
          size: 8,
          shape: 'square',
          strokeWidth: 0,
          fillColors: [COLOR_REQUESTS, COLOR_EARNINGS]
        },
        onItemHover: { highlightDataSeries: true },
        itemMargin: { horizontal: 15, vertical: 8 },
        fontSize: '13px',
        fontWeight: 600
      }
    };
  }, [colorScheme, fontFamily, textPrimary, gridColor, isDark, granularity]);

  return (
    <Box
      sx={{
        '& .apexcharts-toolbar svg': {
          filter: isDark ? 'invert(1) brightness(2)' : 'none'
        }
      }}
    >
      <ReactApexChart options={options} series={series} type="area" height={350} />
    </Box>
  );
}
