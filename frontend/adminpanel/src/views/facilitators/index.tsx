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
import FacilitatorsTable from './FacilitatorsTable';
import axiosServices from 'utils/axios';

// assets
import AddIcon from '@mui/icons-material/AddTwoTone';
import SearchIcon from '@mui/icons-material/Search';

// types
import { Facilitator } from './types';

export default function FacilitatorsPage() {
  const [facilitators, setFacilitators] = useState<Facilitator[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>('');

  useEffect(() => {
    axiosServices
      .get<Facilitator[]>('/api/v1/facilitators')
      .then((res) => setFacilitators(res.data ?? []))
      .catch((err) => setError(err?.error || err?.message || 'Failed to load facilitators'))
      .finally(() => setLoading(false));
  }, []);

  const handleDelete = (id: string) => {
    if (!window.confirm('Delete this facilitator?')) return;
    axiosServices
      .delete(`/api/v1/facilitators/${id}`)
      .then(() => setFacilitators((prev) => prev.filter((f) => f.id !== id)))
      .catch((err) => setError(err?.error || err?.message || 'Failed to delete'));
  };

  const rows = useMemo(() => {
    if (!search.trim()) return facilitators;
    const query = search.toLowerCase();
    return facilitators.filter(
      (f) =>
        f.name.toLowerCase().includes(query) ||
        f.url.toLowerCase().includes(query)
    );
  }, [search, facilitators]);

  return (
    <MainCard content={false}>
      <CardContent>
        <Stack direction={{ xs: 'column', sm: 'row' }} sx={{ alignItems: 'center', justifyContent: 'space-between', gap: 2 }}>
          <TextField
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search Facilitators"
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
          <Button component={Link} to="/facilitators/create" variant="contained" startIcon={<AddIcon fontSize="small" />}>
            Create Facilitator
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
        <FacilitatorsTable rows={rows} onDelete={handleDelete} />
      )}
    </MainCard>
  );
}
