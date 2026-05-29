import { useEffect, useState } from 'react';
import { Link as RouterLink } from 'react-router-dom';

// material-ui
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Chip from '@mui/material/Chip';
import Divider from '@mui/material/Divider';
import Grid from '@mui/material/Grid';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TextField from '@mui/material/TextField';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Typography from '@mui/material/Typography';
import { useTheme } from '@mui/material/styles';

// icons
import AddIcon from '@mui/icons-material/AddTwoTone';
import ArrowUpwardIcon from '@mui/icons-material/ArrowUpward';
import ArrowForwardIcon from '@mui/icons-material/ArrowForward';
import FiberManualRecordIcon from '@mui/icons-material/FiberManualRecord';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import RequestsAndEarnings from 'ui-component/charts/RequestsAndEarnings';
import axiosServices, { axiosProxyServices } from 'utils/axios';
import { gridSpacing } from 'store/constant';

// ─── Types ───────────────────────────────────────────────────────────────────

type PeriodMode = 'day' | 'week' | 'month' | 'custom';

interface ProjectOption {
  id: string;
  name: string;
}

interface PeriodRange {
  from: string;
  to: string;
}

interface PeriodStats {
  range: PeriodRange;
  total_projects: number;
  total_routes: number;
  total_requests: number;
  total_earnings_usd: number;
  success_rate: number;
}

interface DashboardStats {
  period: PeriodStats;
  previous_period: PeriodStats;
}

interface TopRoute {
  path_pattern: string;
  price_usd: string;
  total_requests: number;
  revenue_usd: number;
}

interface RecentRequest {
  id: string;
  path: string;
  method: string;
  created_at: string;
  status_code: number;
  payment_channel: string | null;
  amount_usd: string | null;
}

interface ProxyStatusInfo {
  online: boolean;
  upstream_target: string;
  network: string;
  save_wallet: string;
  asset: string;
  version: string;
  uptime: string;
}

// ─── Mock data ────────────────────────────────────────────────────────────────

const MOCK_PROXY_STATUS: ProxyStatusInfo = {
  online: true,
  upstream_target: 'https://api.example.com',
  network: 'Base Mainnet',
  save_wallet: '0x1a2b...9f0e',
  asset: 'USDC',
  version: 'v1.4.2',
  uptime: '15d 4h 42m'
};


function timeAgo(iso: string): string {
  const diff = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return `${Math.floor(diff / 86400)}d ago`;
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function buildQuery(mode: PeriodMode, from: string, to: string, projectId?: string | null): string {
  let base: string;
  if (mode === 'custom') {
    if (from && to) base = `?from=${from}&to=${to}`;
    else return '';
  } else {
    base = `?period=${mode}`;
  }
  if (projectId) base += `&project_id=${projectId}`;
  return base;
}

function fmtDateRange(from: string, to: string): string {
  const f = new Date(from + 'T00:00:00Z');
  const t = new Date(to + 'T00:00:00Z');
  const months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
  const sameMonth = f.getUTCMonth() === t.getUTCMonth() && f.getUTCFullYear() === t.getUTCFullYear();
  return sameMonth
    ? `${months[f.getUTCMonth()]} ${f.getUTCDate()}–${t.getUTCDate()}, ${t.getUTCFullYear()}`
    : `${months[f.getUTCMonth()]} ${f.getUTCDate()} – ${months[t.getUTCMonth()]} ${t.getUTCDate()}, ${t.getUTCFullYear()}`;
}

interface Delta {
  text: string;
  positive: boolean;
  neutral: boolean;
}

function currencyDelta(cur: number, prev: number): Delta {
  const diff = cur - prev;
  const neutral = diff === 0;
  const positive = diff >= 0;
  const abs = Math.abs(diff).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
  return { text: `${positive && !neutral ? '+' : neutral ? '' : '-'}$${abs} vs prev`, positive, neutral };
}

function percentChangeDelta(cur: number, prev: number): Delta {
  const neutral = cur === prev;
  if (neutral) return { text: 'No change vs prev', positive: true, neutral: true };
  if (prev === 0) {
    const positive = cur > 0;
    return { text: `${positive ? '+' : ''}${cur.toFixed(1)}% vs prev`, positive, neutral: false };
  }
  const pct = ((cur - prev) / prev) * 100;
  const positive = pct >= 0;
  return { text: `${positive ? '+' : ''}${pct.toFixed(1)}% vs prev`, positive, neutral: false };
}

function pointsDelta(cur: number, prev: number): Delta {
  const diff = cur - prev;
  const neutral = diff === 0;
  const positive = diff >= 0;
  if (neutral) return { text: 'No change vs prev', positive: true, neutral: true };
  return { text: `${positive ? '+' : ''}${diff.toFixed(1)}pp vs prev`, positive, neutral: false };
}

function countDelta(cur: number, prev: number): Delta {
  const diff = cur - prev;
  const neutral = diff === 0;
  const positive = diff >= 0;
  if (neutral) return { text: 'No change vs prev', positive: true, neutral: true };
  return { text: `${positive ? '+' : ''}${diff} vs prev`, positive, neutral: false };
}

// ─── Sub-components ───────────────────────────────────────────────────────────

interface StatCardProps {
  label: string;
  value: string | number;
  delta: Delta;
}

function StatCard({ label, value, delta }: StatCardProps) {
  const theme = useTheme();
  const deltaColor = delta.neutral
    ? theme.palette.text.secondary
    : delta.positive
      ? theme.palette.success.main
      : theme.palette.error.main;

  return (
    <Box
      sx={{
        bgcolor: 'background.paper',
        borderRadius: 2,
        p: 2.5,
        height: '100%',
        border: '1px solid',
        borderColor: 'divider',
        transition: 'box-shadow 0.2s',
        '&:hover': { boxShadow: 4 }
      }}
    >
      <Typography variant="caption" color="text.secondary" sx={{ textTransform: 'uppercase', letterSpacing: 0.8, fontWeight: 600 }}>
        {label}
      </Typography>
      <Typography variant="h3" sx={{ mt: 0.5, mb: 1, fontWeight: 700 }}>
        {typeof value === 'number' ? value.toLocaleString() : value}
      </Typography>
      <Stack direction="row" alignItems="center" spacing={0.5}>
        {!delta.neutral && (
          <ArrowUpwardIcon sx={{ fontSize: 14, color: deltaColor, transform: delta.positive ? 'none' : 'rotate(180deg)' }} />
        )}
        <Typography variant="caption" sx={{ color: deltaColor, fontWeight: 500 }}>
          {delta.text}
        </Typography>
      </Stack>
    </Box>
  );
}

interface ProxyStatusRowProps {
  label: string;
  value: string;
  mono?: boolean;
}

function ProxyStatusRow({ label, value, mono = false }: ProxyStatusRowProps) {
  return (
    <Box sx={{ py: 1 }}>
      <Typography variant="caption" color="text.secondary" sx={{ display: 'block', mb: 0.25 }}>
        {label}
      </Typography>
      <Typography variant="body2" sx={{ fontWeight: 500, fontFamily: mono ? 'monospace' : 'inherit', wordBreak: 'break-all' }}>
        {value}
      </Typography>
    </Box>
  );
}

// ─── Period switcher ──────────────────────────────────────────────────────────

interface PeriodSwitcherProps {
  mode: PeriodMode;
  customFrom: string;
  customTo: string;
  onModeChange: (mode: PeriodMode) => void;
  onCustomFromChange: (v: string) => void;
  onCustomToChange: (v: string) => void;
  periodLabel: string | null;
}

function PeriodSwitcher({ mode, customFrom, customTo, onModeChange, onCustomFromChange, onCustomToChange, periodLabel }: PeriodSwitcherProps) {
  return (
    <Stack direction={{ xs: 'column', sm: 'row' }} spacing={1.5} alignItems={{ xs: 'flex-start', sm: 'center' }} flexWrap="wrap">
      <ToggleButtonGroup
        value={mode}
        exclusive
        size="small"
        onChange={(_, v) => { if (v) onModeChange(v as PeriodMode); }}
        sx={{ '& .MuiToggleButton-root': { px: 2, py: 0.5, fontSize: 12, fontWeight: 600, textTransform: 'none' } }}
      >
        <ToggleButton value="day">Day</ToggleButton>
        <ToggleButton value="week">Week</ToggleButton>
        <ToggleButton value="month">Month</ToggleButton>
        <ToggleButton value="custom">Custom</ToggleButton>
      </ToggleButtonGroup>

      {mode === 'custom' && (
        <Stack direction="row" spacing={1} alignItems="center">
          <TextField
            type="date"
            size="small"
            value={customFrom}
            onChange={(e) => onCustomFromChange(e.target.value)}
            slotProps={{ htmlInput: { max: customTo || undefined } }}
            sx={{ width: 150, '& .MuiInputBase-input': { fontSize: 12, py: 0.75 } }}
          />
          <Typography variant="caption" color="text.secondary">—</Typography>
          <TextField
            type="date"
            size="small"
            value={customTo}
            onChange={(e) => onCustomToChange(e.target.value)}
            slotProps={{ htmlInput: { min: customFrom || undefined } }}
            sx={{ width: 150, '& .MuiInputBase-input': { fontSize: 12, py: 0.75 } }}
          />
        </Stack>
      )}

      {periodLabel && mode !== 'custom' && (
        <Chip label={periodLabel} size="small" variant="outlined" sx={{ fontSize: 11, height: 22, borderRadius: 1 }} />
      )}
    </Stack>
  );
}

// ─── Dashboard page ───────────────────────────────────────────────────────────

interface GatewayHealth {
  online: boolean;
  latency_ms?: number;
  error?: string;
  status?: string;
}

export default function DashboardPage() {
  const [stats, setStats]               = useState<DashboardStats | null>(null);
  const [gatewayHealth, setGatewayHealth] = useState<GatewayHealth | null>(null);
  const [periodMode, setPeriodMode]     = useState<PeriodMode>('week');
  const [customFrom, setCustomFrom]     = useState('');
  const [customTo, setCustomTo]         = useState('');
  const [projects, setProjects]         = useState<ProjectOption[]>([]);
  const [selectedProject, setSelectedProject] = useState<string>('');
  const [recentRequests, setRecentRequests] = useState<RecentRequest[]>([]);
  const [topRoutes, setTopRoutes] = useState<TopRoute[]>([]);

  useEffect(() => {
    axiosServices
      .get<{ id: string; name: string }[]>('/api/v1/projects')
      .then((res) => setProjects(res.data.map((p) => ({ id: p.id, name: p.name }))))
      .catch(() => {});
  }, []);

  const projectId = selectedProject || null;

  useEffect(() => {
    const query = buildQuery(periodMode, customFrom, customTo, projectId);
    if (!query) return; // custom mode but dates not filled yet
    axiosServices
      .get<DashboardStats>(`/api/v1/stats/dashboard${query}`)
      .then((res) => setStats(res.data))
      .catch(() => {});
  }, [periodMode, customFrom, customTo, projectId]);

  useEffect(() => {
    const qs = projectId ? `?project_id=${projectId}` : '';
    axiosServices
      .get<RecentRequest[]>(`/api/v1/stats/recent-requests${qs}`)
      .then((res) => setRecentRequests(res.data))
      .catch(() => {});
    axiosServices
      .get<TopRoute[]>(`/api/v1/stats/top-routes${qs}`)
      .then((res) => setTopRoutes(res.data))
      .catch(() => {});
  }, [projectId]);

  useEffect(() => {
    const check = () => {
      const start = Date.now();
      axiosProxyServices
        .get<GatewayHealth>('/healthz')
        .then((res) => setGatewayHealth({ online: res.data?.status === 'ok', latency_ms: Date.now() - start }))
        .catch(() => setGatewayHealth({ online: false, error: 'unreachable' }));
    };
    check();
    const interval = setInterval(check, 5_000);
    return () => clearInterval(interval);
  }, []);

  const p  = stats?.period;
  const pp = stats?.previous_period;

  const earningsFormatted = `$${(p?.total_earnings_usd ?? 0).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
  const periodLabel       = p ? fmtDateRange(p.range.from, p.range.to) : null;

  const revenueD  = p && pp ? currencyDelta(p.total_earnings_usd, pp.total_earnings_usd) : { text: '—', positive: true, neutral: true };
  const requestsD = p && pp ? percentChangeDelta(p.total_requests, pp.total_requests)    : { text: '—', positive: true, neutral: true };
  const successD  = p && pp ? pointsDelta(p.success_rate, pp.success_rate)               : { text: '—', positive: true, neutral: true };
  const routesD   = p && pp ? countDelta(p.total_routes, pp.total_routes)                : { text: '—', positive: true, neutral: true };

  return (
    <Grid container spacing={gridSpacing}>

      {/* ── Page header ── */}
      <Grid size={{ xs: 12 }}>
        <MainCard border boxShadow sx={{ py: 0.5 }}>
          <Stack spacing={2}>
            <Stack direction={{ xs: 'column', sm: 'row' }} justifyContent="space-between" alignItems={{ xs: 'flex-start', sm: 'center' }} spacing={2}>
              <Box>
                <Typography variant="h2" sx={{ mb: 0.25 }}>Dashboard</Typography>
                <Typography variant="body2" color="text.secondary">
                  Monitor your protected API routes, payments and proxy traffic.
                </Typography>
              </Box>
              <Stack direction="row" spacing={1.5} alignItems="center" flexShrink={0}>
                <Chip
                  icon={<FiberManualRecordIcon sx={{ fontSize: '10px !important', color: gatewayHealth === null ? 'text.disabled' : gatewayHealth.online ? 'success.main' : 'error.main' }} />}
                  label={gatewayHealth === null ? 'Checking…' : gatewayHealth.online ? 'Live' : 'Offline'}
                  size="small"
                  variant="outlined"
                  sx={{
                    borderColor: gatewayHealth === null ? 'divider' : gatewayHealth.online ? 'success.main' : 'error.main',
                    color:       gatewayHealth === null ? 'text.disabled' : gatewayHealth.online ? 'success.main' : 'error.main',
                    fontWeight: 600
                  }}
                />
                <Button component={RouterLink} to="/routes/create" variant="contained" size="small" startIcon={<AddIcon />}>
                  Create Route
                </Button>
              </Stack>
            </Stack>

            <Divider />

            <Stack direction={{ xs: 'column', sm: 'row' }} spacing={2} alignItems={{ xs: 'flex-start', sm: 'center' }} flexWrap="wrap">
              <Select
                size="small"
                value={selectedProject}
                onChange={(e) => setSelectedProject(e.target.value)}
                displayEmpty
                sx={{ minWidth: 180, fontSize: 13, '& .MuiSelect-select': { py: 0.75 } }}
              >
                <MenuItem value="">All Projects</MenuItem>
                {projects.map((p) => (
                  <MenuItem key={p.id} value={p.id}>{p.name}</MenuItem>
                ))}
              </Select>

              <PeriodSwitcher
                mode={periodMode}
                customFrom={customFrom}
                customTo={customTo}
                onModeChange={setPeriodMode}
                onCustomFromChange={setCustomFrom}
                onCustomToChange={setCustomTo}
                periodLabel={periodLabel}
              />
            </Stack>
          </Stack>
        </MainCard>
      </Grid>

      {/* ── Stat cards ── */}
      <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
        <StatCard label="Revenue"        value={earningsFormatted}         delta={revenueD}  />
      </Grid>
      <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
        <StatCard label="Total Requests" value={p?.total_requests ?? 0}    delta={requestsD} />
      </Grid>
      <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
        <StatCard
          label="Success Rate"
          value={p ? `${p.success_rate.toFixed(1)}%` : '—'}
          delta={successD}
        />
      </Grid>
      <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
        <StatCard label="Active Routes"  value={p?.total_routes ?? 0}      delta={routesD}   />
      </Grid>

      {/* ── Chart + Proxy Status ── */}
      <Grid size={{ xs: 12 }}>
        <MainCard border boxShadow title="Traffic & Revenue" sx={{ height: '100%' }}>
          <RequestsAndEarnings periodMode={periodMode} customFrom={customFrom} customTo={customTo} projectId={projectId} />
        </MainCard>
      </Grid>

      {/*<Grid size={{ xs: 12, lg: 4 }}>*/}
      {/*  <MainCard*/}
      {/*    border*/}
      {/*    boxShadow*/}
      {/*    title="Proxy Status"*/}
      {/*    secondary={*/}
      {/*      <Chip*/}
      {/*        icon={<FiberManualRecordIcon sx={{ fontSize: '10px !important' }} />}*/}
      {/*        label={MOCK_PROXY_STATUS.online ? 'Online' : 'Offline'}*/}
      {/*        size="small"*/}
      {/*        color={MOCK_PROXY_STATUS.online ? 'success' : 'error'}*/}
      {/*      />*/}
      {/*    }*/}
      {/*    sx={{ height: '100%' }}*/}
      {/*  >*/}
      {/*    <Stack divider={<Divider flexItem />}>*/}
      {/*      <ProxyStatusRow label="Upstream Target" value={MOCK_PROXY_STATUS.upstream_target} mono />*/}
      {/*      <ProxyStatusRow label="Network"         value={MOCK_PROXY_STATUS.network} />*/}
      {/*      <ProxyStatusRow label="Safe Wallet"     value={MOCK_PROXY_STATUS.save_wallet} mono />*/}
      {/*      <ProxyStatusRow label="Asset"           value={MOCK_PROXY_STATUS.asset} />*/}
      {/*      <ProxyStatusRow label="Version"         value={MOCK_PROXY_STATUS.version} />*/}
      {/*      <ProxyStatusRow label="Uptime"          value={MOCK_PROXY_STATUS.uptime} />*/}
      {/*    </Stack>*/}
      {/*  </MainCard>*/}
      {/*</Grid>*/}

      {/* ── Top Routes + Recent Requests ── */}
      <Grid size={{ xs: 12, lg: 6 }}>
        <MainCard
          border
          boxShadow
          title="Top Routes"
          secondary={
            <Button component={RouterLink} to="/routes" size="small" endIcon={<ArrowForwardIcon />} sx={{ fontSize: 12 }}>
              View all routes
            </Button>
          }
          contentSX={{ p: 0, '&:last-child': { pb: 0 } }}
        >
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Route</TableCell>
                  <TableCell align="right">Price (USDC)</TableCell>
                  <TableCell align="right">Requests</TableCell>
                  <TableCell align="right">Revenue (USDC)</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {topRoutes.map((row) => (
                  <TableRow key={row.path_pattern} hover>
                    <TableCell sx={{ fontFamily: 'monospace', fontSize: 12 }}>{row.path_pattern}</TableCell>
                    <TableCell align="right">{row.price_usd ? `$${row.price_usd}` : '—'}</TableCell>
                    <TableCell align="right">{row.total_requests.toLocaleString()}</TableCell>
                    <TableCell align="right">${row.revenue_usd.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </MainCard>
      </Grid>

      <Grid size={{ xs: 12, lg: 6 }}>
        <MainCard
          border
          boxShadow
          title="Recent Requests"
          secondary={
            <Button component={RouterLink} to="/stats" size="small" endIcon={<ArrowForwardIcon />} sx={{ fontSize: 12 }}>
              View all
            </Button>
          }
          contentSX={{ p: 0, '&:last-child': { pb: 0 } }}
        >
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Time</TableCell>
                  <TableCell>Route</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Payment</TableCell>
                  <TableCell align="right">Amount</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {recentRequests.map((row) => (
                  <TableRow key={row.id} hover>
                    <TableCell sx={{ color: 'text.secondary', fontSize: 12, whiteSpace: 'nowrap' }}>{timeAgo(row.created_at)}</TableCell>
                    <TableCell sx={{ fontFamily: 'monospace', fontSize: 12 }}>{row.path}</TableCell>
                    <TableCell>
                      <Chip label={row.status_code} size="small" color={row.status_code === 200 ? 'success' : 'warning'} sx={{ fontWeight: 700, fontSize: 11, height: 20 }} />
                    </TableCell>
                    <TableCell sx={{ fontSize: 12 }}>{row.payment_channel ?? '—'}</TableCell>
                    <TableCell align="right" sx={{ fontFamily: 'monospace', fontSize: 12 }}>
                      {row.amount_usd ? `$${row.amount_usd}` : '—'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </MainCard>
      </Grid>

    </Grid>
  );
}
