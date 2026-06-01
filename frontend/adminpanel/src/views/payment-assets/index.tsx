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
import PaymentAssetsTable from './PaymentAssetsTable';
import axiosServices from 'utils/axios';

// assets
import AddIcon from '@mui/icons-material/AddTwoTone';
import SearchIcon from '@mui/icons-material/Search';

// types
import { PaymentMethodAsset } from './types';

export default function PaymentAssetsPage() {
  const [assets, setAssets] = useState<PaymentMethodAsset[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState<string>('');

  useEffect(() => {
    axiosServices
      .get<PaymentMethodAsset[]>('/api/v1/payment-method-assets')
      .then((res) => setAssets(res.data ?? []))
      .catch((err) => setError(err?.error || err?.message || 'Failed to load payment assets'))
      .finally(() => setLoading(false));
  }, []);

  const handleDelete = (id: string) => {
    if (!window.confirm('Delete this payment asset?')) return;
    axiosServices
      .delete(`/api/v1/payment-method-assets/${id}`)
      .then(() => setAssets((prev) => prev.filter((a) => a.id !== id)))
      .catch((err) => setError(err?.error || err?.message || 'Failed to delete'));
  };

  const rows = useMemo(() => {
    if (!search.trim()) return assets;
    const query = search.toLowerCase();
    return assets.filter(
      (a) =>
        a.symbol.toLowerCase().includes(query) ||
        a.payment_method_name.toLowerCase().includes(query) ||
        (a.payment_method_chain ?? '').toLowerCase().includes(query) ||
        (a.contract_address ?? '').toLowerCase().includes(query)
    );
  }, [search, assets]);

  return (
    <MainCard content={false}>
      <CardContent>
        <Stack direction={{ xs: 'column', sm: 'row' }} sx={{ alignItems: 'center', justifyContent: 'space-between', gap: 2 }}>
          <TextField
            value={search}
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search Payment Assets"
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
          <Button component={Link} to="/payment-assets/create" variant="contained" startIcon={<AddIcon fontSize="small" />}>
            Create Payment Asset
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
        <PaymentAssetsTable rows={rows} onDelete={handleDelete} />
      )}
    </MainCard>
  );
}
