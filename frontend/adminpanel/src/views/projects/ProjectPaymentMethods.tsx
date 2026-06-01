import * as React from 'react';
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Chip from '@mui/material/Chip';
import CircularProgress from '@mui/material/CircularProgress';
import Dialog from '@mui/material/Dialog';
import DialogActions from '@mui/material/DialogActions';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import IconButton from '@mui/material/IconButton';
import InputLabel from '@mui/material/InputLabel';
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
import Typography from '@mui/material/Typography';
import AddIcon from '@mui/icons-material/Add';
import DeleteTwoToneIcon from '@mui/icons-material/DeleteTwoTone';
import EditTwoToneIcon from '@mui/icons-material/EditTwoTone';
import { Formik } from 'formik';
import * as Yup from 'yup';

import axios from 'utils/axios';
import { PaymentMethod } from '../payment-methods/types';
import { PaymentMethodAsset } from '../payment-assets/types';
import { Facilitator } from '../facilitators/types';
import { ProjectPaymentMethod } from './types';

const SCHEMES = ['exact', 'upto', 'batch-payment'] as const;

interface Props {
  projectId: string;
  isView: boolean;
  canEdit: boolean;
}

const emptyForm = {
  payment_method_id: '',
  asset_id: '',
  scheme: 'exact' as string,
  facilitator_id: '',
  payout_address: '',
  enabled: true,
};

export default function ProjectPaymentMethods({ projectId, isView, canEdit }: Props) {
  const showActions = !isView && canEdit;
  const [methods, setMethods] = React.useState<ProjectPaymentMethod[]>([]);
  const [paymentMethods, setPaymentMethods] = React.useState<PaymentMethod[]>([]);
  const [assets, setAssets] = React.useState<PaymentMethodAsset[]>([]);
  const [facilitators, setFacilitators] = React.useState<Facilitator[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [dialogOpen, setDialogOpen] = React.useState(false);
  const [editing, setEditing] = React.useState<ProjectPaymentMethod | null>(null);

  const loadMethods = React.useCallback(async () => {
    const res = await axios.get<ProjectPaymentMethod[]>(`/api/v1/project-payment-methods?project_id=${projectId}`);
    setMethods(res.data);
  }, [projectId]);

  React.useEffect(() => {
    Promise.all([
      loadMethods(),
      axios.get<PaymentMethod[]>('/api/v1/payment-methods').then(r => setPaymentMethods(r.data)),
      axios.get<PaymentMethodAsset[]>('/api/v1/payment-method-assets').then(r => setAssets(r.data)),
      axios.get<Facilitator[]>('/api/v1/facilitators').then(r => setFacilitators(r.data)),
    ]).finally(() => setLoading(false));
  }, [loadMethods]);

  const handleDelete = async (id: string) => {
    if (!window.confirm('Delete this payment method configuration?')) return;
    await axios.delete(`/api/v1/project-payment-methods/${id}`);
    await loadMethods();
  };

  const handleDialogClose = () => {
    setDialogOpen(false);
    setEditing(null);
  };

  const handleEdit = (m: ProjectPaymentMethod) => {
    setEditing(m);
    setDialogOpen(true);
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  const getMethodName = (id: string) => paymentMethods.find(x => x.id === id)?.name ?? id;
  const getAssetLabel = (id: string) => assets.find(x => x.id === id)?.symbol ?? id;
  const getFacilitatorName = (id: string) => facilitators.find(x => x.id === id)?.name ?? id;

  return (
    <Box>
      {showActions && (
        <Box sx={{ mb: 2, display: 'flex', justifyContent: 'flex-end' }}>
          <Button variant="contained" startIcon={<AddIcon />} onClick={() => setDialogOpen(true)}>
            Add Payment Method
          </Button>
        </Box>
      )}

      {methods.length === 0 ? (
        <Typography color="text.secondary" sx={{ py: 4, textAlign: 'center' }}>
          No payment methods configured for this project.
        </Typography>
      ) : (
        <TableContainer>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Method</TableCell>
                <TableCell>Asset</TableCell>
                <TableCell>Scheme</TableCell>
                <TableCell>Facilitator</TableCell>
                <TableCell>Payout Address</TableCell>
                <TableCell>Status</TableCell>
                {showActions && <TableCell align="right">Actions</TableCell>}
              </TableRow>
            </TableHead>
            <TableBody>
              {methods.map((m) => (
                <TableRow key={m.id} hover>
                  <TableCell>{getMethodName(m.payment_method_id)}</TableCell>
                  <TableCell>{getAssetLabel(m.asset_id)}</TableCell>
                  <TableCell>
                    <Chip label={m.scheme} size="small" variant="outlined" />
                  </TableCell>
                  <TableCell>{getFacilitatorName(m.facilitator_id)}</TableCell>
                  <TableCell sx={{ fontFamily: 'monospace', fontSize: 12, maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {m.payout_address ?? '—'}
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={m.enabled ? 'Enabled' : 'Disabled'}
                      size="small"
                      color={m.enabled ? 'success' : 'default'}
                    />
                  </TableCell>
                  {showActions && (
                    <TableCell align="right">
                      <IconButton size="small" onClick={() => handleEdit(m)}>
                        <EditTwoToneIcon fontSize="small" />
                      </IconButton>
                      <IconButton size="small" color="error" onClick={() => handleDelete(m.id)}>
                        <DeleteTwoToneIcon fontSize="small" />
                      </IconButton>
                    </TableCell>
                  )}
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}

      <Dialog open={dialogOpen} onClose={handleDialogClose} maxWidth="sm" fullWidth>
        <Formik
          enableReinitialize
          initialValues={
            editing
              ? {
                  payment_method_id: editing.payment_method_id,
                  asset_id: editing.asset_id,
                  scheme: editing.scheme,
                  facilitator_id: editing.facilitator_id,
                  payout_address: editing.payout_address ?? '',
                  enabled: editing.enabled,
                }
              : emptyForm
          }
          validationSchema={Yup.object().shape({
            payment_method_id: Yup.string().required('Payment method is required'),
            asset_id: Yup.string().required('Asset is required'),
            scheme: Yup.string().required('Scheme is required'),
            facilitator_id: Yup.string().required('Facilitator is required'),
          })}
          onSubmit={async (values, { setErrors, setSubmitting }) => {
            try {
              if (editing) {
                await axios.put(`/api/v1/project-payment-methods/${editing.id}`, {
                  scheme: values.scheme,
                  facilitator_id: values.facilitator_id,
                  payout_address: values.payout_address || null,
                  enabled: values.enabled,
                });
              } else {
                await axios.post('/api/v1/project-payment-methods', {
                  project_id: projectId,
                  payment_method_id: values.payment_method_id,
                  asset_id: values.asset_id,
                  scheme: values.scheme,
                  facilitator_id: values.facilitator_id,
                  payout_address: values.payout_address || null,
                  enabled: values.enabled,
                });
              }
              await loadMethods();
              handleDialogClose();
            } catch (err: any) {
              setErrors({ payment_method_id: err?.error || err?.message || 'Request failed' });
              setSubmitting(false);
            }
          }}
        >
          {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, setFieldValue, touched, values }) => {
            const filteredAssets = assets.filter((a) => a.payment_method_id === values.payment_method_id);

            return (
              <form onSubmit={handleSubmit}>
                <DialogTitle>{editing ? 'Edit Payment Method' : 'Add Payment Method'}</DialogTitle>
                <DialogContent>
                  <Stack spacing={2.5} sx={{ mt: 1 }}>
                    <FormControl fullWidth error={Boolean(touched.payment_method_id && errors.payment_method_id)}>
                      <InputLabel>Payment Method</InputLabel>
                      <Select
                        name="payment_method_id"
                        value={values.payment_method_id}
                        label="Payment Method"
                        disabled={!!editing}
                        onChange={(e) => {
                          handleChange(e);
                          setFieldValue('asset_id', '');
                        }}
                        onBlur={handleBlur}
                      >
                        {paymentMethods.map((pm) => (
                          <MenuItem key={pm.id} value={pm.id}>
                            {pm.name} ({pm.protocol})
                          </MenuItem>
                        ))}
                      </Select>
                      {touched.payment_method_id && errors.payment_method_id && (
                        <FormHelperText>{errors.payment_method_id}</FormHelperText>
                      )}
                    </FormControl>

                    <FormControl fullWidth error={Boolean(touched.asset_id && errors.asset_id)}>
                      <InputLabel>Asset</InputLabel>
                      <Select
                        name="asset_id"
                        value={values.asset_id}
                        label="Asset"
                        disabled={!!editing || !values.payment_method_id}
                        onChange={handleChange}
                        onBlur={handleBlur}
                      >
                        {filteredAssets.map((a) => (
                          <MenuItem key={a.id} value={a.id}>
                            {a.symbol}
                          </MenuItem>
                        ))}
                      </Select>
                      {touched.asset_id && errors.asset_id && (
                        <FormHelperText>{errors.asset_id}</FormHelperText>
                      )}
                    </FormControl>

                    <FormControl fullWidth>
                      <InputLabel>Scheme</InputLabel>
                      <Select
                        name="scheme"
                        value={values.scheme}
                        label="Scheme"
                        onChange={handleChange}
                        onBlur={handleBlur}
                      >
                        {SCHEMES.map((s) => (
                          <MenuItem key={s} value={s}>
                            {s}
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>

                    <FormControl fullWidth error={Boolean(touched.facilitator_id && errors.facilitator_id)}>
                      <InputLabel>Facilitator</InputLabel>
                      <Select
                        name="facilitator_id"
                        value={values.facilitator_id}
                        label="Facilitator"
                        onChange={handleChange}
                        onBlur={handleBlur}
                      >
                        {facilitators.map((f) => (
                          <MenuItem key={f.id} value={f.id}>
                            {f.name}
                          </MenuItem>
                        ))}
                      </Select>
                      {touched.facilitator_id && errors.facilitator_id && (
                        <FormHelperText>{errors.facilitator_id}</FormHelperText>
                      )}
                    </FormControl>

                    <TextField
                      fullWidth
                      label="Payout Address"
                      name="payout_address"
                      value={values.payout_address}
                      onChange={handleChange}
                      onBlur={handleBlur}
                      helperText="Optional. Wallet address for receiving payments in this network."
                    />

                    <FormControlLabel
                      control={
                        <Checkbox
                          name="enabled"
                          checked={values.enabled}
                          onChange={handleChange}
                          color="primary"
                        />
                      }
                      label="Enabled"
                      sx={{ ml: 0 }}
                    />
                  </Stack>
                </DialogContent>
                <DialogActions>
                  <Button onClick={handleDialogClose} disabled={isSubmitting}>
                    Cancel
                  </Button>
                  <Button type="submit" variant="contained" disabled={isSubmitting}>
                    {editing ? 'Save Changes' : 'Add'}
                  </Button>
                </DialogActions>
              </form>
            );
          }}
        </Formik>
      </Dialog>
    </Box>
  );
}
