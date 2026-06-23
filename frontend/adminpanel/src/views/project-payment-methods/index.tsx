import * as React from 'react';
import Alert from '@mui/material/Alert';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import Typography from '@mui/material/Typography';

import MainCard from 'ui-component/cards/MainCard';
import axios from 'utils/axios';

interface ProjectPaymentMethodFull {
  id: string;
  project_id: string;
  project_name: string;
  payment_method_id: string;
  payment_method_name: string;
  asset_id: string;
  asset_symbol: string;
  scheme: string;
  // x402 only — omitted for MPP links (control-api returns them nullable).
  facilitator_id?: string | null;
  facilitator_name?: string | null;
  payout_address?: string | null;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export default function ProjectPaymentMethodsPage() {
  const [rows, setRows] = React.useState<ProjectPaymentMethodFull[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState('');

  React.useEffect(() => {
    axios
      .get<ProjectPaymentMethodFull[]>('/api/v1/project-payment-methods/all')
      .then((r) => setRows(r.data))
      .catch(() => setError('Failed to load data'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) {
    return (
      <MainCard title="Project Payment Methods">
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      </MainCard>
    );
  }

  if (error) {
    return (
      <MainCard title="Project Payment Methods">
        <Typography color="error">{error}</Typography>
      </MainCard>
    );
  }

  return (
    <MainCard title="Project Payment Methods">
      <Alert severity="info" sx={{ mb: 2 }}>
        To add or modify a payment method for a project, go to the project&apos;s edit page.
      </Alert>
      {rows.length === 0 ? (
        <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>
          No project payment methods configured.
        </Typography>
      ) : (
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Project</TableCell>
                <TableCell>Method</TableCell>
                <TableCell>Asset</TableCell>
                <TableCell>Scheme</TableCell>
                <TableCell>Facilitator</TableCell>
                <TableCell>Payout Address</TableCell>
                <TableCell>Status</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {rows.map((m) => (
                <TableRow key={m.id} hover>
                  <TableCell>{m.project_name}</TableCell>
                  <TableCell>{m.payment_method_name}</TableCell>
                  <TableCell>{m.asset_symbol}</TableCell>
                  <TableCell>
                    <Chip label={m.scheme} size="small" variant="outlined" />
                  </TableCell>
                  <TableCell>{m.facilitator_name ?? '—'}</TableCell>
                  <TableCell sx={{ fontFamily: 'monospace', fontSize: 12, maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {m.payout_address ?? '—'}
                  </TableCell>
                  <TableCell>
                    <Chip label={m.enabled ? 'Enabled' : 'Disabled'} size="small" color={m.enabled ? 'success' : 'default'} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </MainCard>
  );
}
