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
import PaymentChannelsTable from './PaymentChannelsTable';
import axiosServices from 'utils/axios';

// assets
import AddIcon from '@mui/icons-material/AddTwoTone';
import SearchIcon from '@mui/icons-material/Search';

// types
import { PaymentChannel } from './types';

export default function PaymentChannelsPage() {
  const [channels, setChannels] = useState<PaymentChannel[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>('');

  useEffect(() => {
    axiosServices
      .get<PaymentChannel[]>('/api/v1/payment-channels')
      .then((res) => setChannels(res.data ?? []))
      .catch((err) => setError(err?.error || err?.message || 'Failed to load payment channels'))
      .finally(() => setLoading(false));
  }, []);

  const rows = useMemo(() => {
    if (!search.trim()) return channels;
    const query = search.toLowerCase();
    return channels.filter(
      (c) =>
        c.protocol.toLowerCase().includes(query) ||
        c.method.toLowerCase().includes(query) ||
        c.scheme.toLowerCase().includes(query)
    );
  }, [search, channels]);

  return (
    <MainCard content={false}>
      <CardContent>
        <Stack direction={{ xs: 'column', sm: 'row' }} sx={{ alignItems: 'center', justifyContent: 'space-between', gap: 2 }}>
          <TextField
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search Payment Channel"
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

          <Button component={Link} to="/payment-channels/create" variant="contained" startIcon={<AddIcon fontSize="small" />}>
            Create Payment Channel
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
        <PaymentChannelsTable rows={rows} />
      )}
    </MainCard>
  );
}
