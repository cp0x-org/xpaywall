import { useEffect, useMemo, useState } from 'react';

// material-ui
import Alert from '@mui/material/Alert';
import CardContent from '@mui/material/CardContent';
import CircularProgress from '@mui/material/CircularProgress';
import InputAdornment from '@mui/material/InputAdornment';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import RequestsTable from './RequestsTable';
import axiosServices from 'utils/axios';

// assets
import SearchIcon from '@mui/icons-material/Search';

// types
import { RequestLog } from './types';

export default function RequestsPage() {
  const [logs, setLogs] = useState<RequestLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');

  useEffect(() => {
    axiosServices
      .get<RequestLog[]>('/api/v1/request-logs?limit=200')
      .then((res) => setLogs(res.data ?? []))
      .catch((err) => setError(err?.message || 'Failed to load request logs'))
      .finally(() => setLoading(false));
  }, []);

  const rows = useMemo(() => {
    if (!search.trim()) return logs;
    const q = search.toLowerCase();
    return logs.filter(
      (r) =>
        r.path.toLowerCase().includes(q) ||
        r.method.toLowerCase().includes(q) ||
        (r.client_ip ?? '').toLowerCase().includes(q) ||
        (r.status ?? '').toLowerCase().includes(q)
    );
  }, [search, logs]);

  return (
    <MainCard border boxShadow content={false}>
      <CardContent>
        <Stack direction={{ xs: 'column', sm: 'row' }} sx={{ alignItems: 'center', justifyContent: 'space-between', gap: 2 }}>
          <Typography variant="h3">Request Logs</Typography>
          <TextField
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by path, method, IP…"
            size="small"
            sx={{ width: { xs: 1, sm: 320 } }}
            slotProps={{
              input: {
                startAdornment: (
                  <InputAdornment position="start">
                    <SearchIcon fontSize="small" />
                  </InputAdornment>
                )
              }
            }}
          />
        </Stack>
      </CardContent>

      {error && (
        <CardContent>
          <Alert severity="error">{error}</Alert>
        </CardContent>
      )}

      {loading ? (
        <CardContent sx={{ display: 'flex', justifyContent: 'center', py: 6 }}>
          <CircularProgress />
        </CardContent>
      ) : (
        <RequestsTable rows={rows} />
      )}
    </MainCard>
  );
}
