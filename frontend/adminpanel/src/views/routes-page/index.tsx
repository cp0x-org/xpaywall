import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';

// material-ui
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import CardContent from '@mui/material/CardContent';
import CircularProgress from '@mui/material/CircularProgress';
import InputAdornment from '@mui/material/InputAdornment';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import RoutesTable from './RoutesTable';
import axiosServices from 'utils/axios';

// assets
import AddIcon from '@mui/icons-material/AddTwoTone';
import SearchIcon from '@mui/icons-material/Search';

// types
import { RouteRow } from './types';
import { Project, ProxyUrl } from '../projects/types';

export default function RoutesPage() {
  const [routes, setRoutes] = useState<RouteRow[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>('');
  const [proxyUrl, setProxyUrl] = useState<string>('');
  const [projectBaseUrls, setProjectBaseUrls] = useState<Record<string, string>>({});

  useEffect(() => {
    const fetchRoutes = axiosServices
      .get<RouteRow[]>('/api/v1/outbound-routes')
      .then((res) => setRoutes(res.data ?? []));

    const fetchProxy = axiosServices
      .get<ProxyUrl>('/api/v1/system/proxy-url')
      .then((res) => setProxyUrl(res.data.proxy_url ?? ''));

    const fetchProjects = axiosServices
      .get<Project[]>('/api/v1/projects/with-config')
      .then((res) => {
        const map: Record<string, string> = {};
        for (const p of res.data ?? []) map[p.id] = p.base_url ?? '';
        setProjectBaseUrls(map);
      });

    Promise.all([fetchRoutes, fetchProxy, fetchProjects])
      .catch((err) => setError(err?.error || err?.message || 'Failed to load routes'))
      .finally(() => setLoading(false));
  }, []);

  const rows = useMemo(() => {
    if (!search.trim()) return routes;
    const query = search.toLowerCase();
    return routes.filter(
      (r) =>
        r.name.toLowerCase().includes(query) ||
        r.path_pattern.toLowerCase().includes(query) ||
        r.id.includes(query)
    );
  }, [search, routes]);

  return (
    <MainCard content={false}>
      <CardContent>
        <Stack direction={{ xs: 'column', sm: 'row' }} sx={{ alignItems: 'center', justifyContent: 'space-between', gap: 2 }}>
          <TextField
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search Route"
            size="small"
            sx={{ width: { xs: 1, sm: 'auto' } }}
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

          <Button component={Link} to="/routes/create" variant="contained" startIcon={<AddIcon fontSize="small" />}>
            Create Route
          </Button>
        </Stack>
      </CardContent>

      {error && (
        <CardContent>
          <Alert severity="error">{error}</Alert>
        </CardContent>
      )}

      {loading ? (
        <CardContent sx={{ display: 'flex', justifyContent: 'center', py: 4 }}>
          <CircularProgress />
        </CardContent>
      ) : (
        <RoutesTable rows={rows} proxyUrl={proxyUrl} projectBaseUrls={projectBaseUrls} />
      )}
    </MainCard>
  );
}
