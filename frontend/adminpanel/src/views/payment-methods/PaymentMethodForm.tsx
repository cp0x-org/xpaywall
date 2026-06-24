import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

// material-ui
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import CircularProgress from '@mui/material/CircularProgress';
import EditTwoToneIcon from '@mui/icons-material/EditTwoTone';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import ToggleButton from '@mui/material/ToggleButton';
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup';
import Typography from '@mui/material/Typography';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import GlobalScopeToggle from 'ui-component/GlobalScopeToggle';
import axios from 'utils/axios';
import { PaymentMethod } from './types';

interface NetworkItem {
  caip2: string;
  name: string;
}

// MPP-only option sets. x402 methods use a facilitator + network instead.
const METHODS = ['tempo', 'stripe'] as const;
const SCHEMES = ['charge'] as const;

const emptyValues = {
  code: '',
  protocol: 'x402',
  name: '',
  caip2_chain_id: '',
  method: 'tempo',
  scheme: 'charge',
  enabled: true,
  is_global: false,
  submit: null
};

export default function PaymentMethodForm() {
  const { pathname, state } = useLocation();
  const navigate = useNavigate();

  const isCreate = pathname.includes('/create');
  const isEdit = pathname.includes('/edit');
  const isView = pathname.includes('/view');

  const id: string | undefined = (state as any)?.id;

  const [loading, setLoading] = useState(isEdit || isView);
  const [loadError, setLoadError] = useState('');
  const [initialValues, setInitialValues] = useState(emptyValues);
  const [networks, setNetworks] = useState<NetworkItem[]>([]);
  const [networkMode, setNetworkMode] = useState<'select' | 'custom'>('select');

  useEffect(() => {
    axios
      .get<NetworkItem[]>('/api/v1/system/networks')
      .then((res) => setNetworks(res.data))
      .catch(() => {});
  }, []);

  useEffect(() => {
    if (!(isEdit || isView) || networks.length === 0) return;
    const caip2 = initialValues.caip2_chain_id;
    if (caip2 && !networks.find((n) => n.caip2 === caip2)) {
      setNetworkMode('custom');
    }
  }, [networks, initialValues.caip2_chain_id, isEdit, isView]);

  useEffect(() => {
    if ((isEdit || isView) && id) {
      axios
        .get<PaymentMethod>(`/api/v1/payment-methods/${id}`)
        .then((res) => {
          const d = res.data;
          setInitialValues({
            code: d.code,
            protocol: d.protocol,
            name: d.name,
            caip2_chain_id: d.caip2_chain_id ?? '',
            method: d.method ?? 'tempo',
            scheme: d.scheme ?? 'charge',
            enabled: d.enabled,
            is_global: d.is_global ?? false,
            submit: null
          });
        })
        .catch(() => setLoadError('Failed to load payment method'))
        .finally(() => setLoading(false));
    }
  }, [isEdit, isView, id]);

  let title = 'Payment Method';
  if (isCreate) title = 'Create Payment Method';
  if (isEdit) title = 'Update Payment Method';
  if (isView) title = 'View Payment Method';

  if (loading) {
    return (
      <MainCard title={title}>
        <Box sx={{ display: 'flex', justifyContent: 'center', p: 4 }}>
          <CircularProgress />
        </Box>
      </MainCard>
    );
  }

  if (loadError) {
    return (
      <MainCard title={title}>
        <Typography color="error">{loadError}</Typography>
      </MainCard>
    );
  }

  return (
    <MainCard title={title}>
      <Formik
        enableReinitialize
        initialValues={initialValues}
        validationSchema={Yup.object().shape({
          code: Yup.string().required('Code is required'),
          protocol: Yup.string().required('Protocol is required'),
          // MPP doesn't need a name; x402 derives one from the network or custom entry.
          name: Yup.string().when('protocol', {
            is: 'mpp',
            then: (s) => s,
            otherwise: (s) => s.required('Name is required')
          }),
          method: Yup.string().when('protocol', {
            is: 'mpp',
            then: (s) => s.required('Method is required')
          }),
          scheme: Yup.string().when('protocol', {
            is: 'mpp',
            then: (s) => s.required('Scheme is required')
          })
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          try {
            const isMPP = values.protocol === 'mpp';
            const payload = {
              code: values.code,
              protocol: values.protocol,
              name: values.name,
              // x402 settles via a facilitator + network; MPP via method/scheme.
              caip2_chain_id: isMPP ? null : values.caip2_chain_id || null,
              method: isMPP ? values.method : null,
              scheme: isMPP ? values.scheme : null,
              enabled: values.enabled,
              is_global: values.is_global
            };
            if (isCreate) {
              await axios.post('/api/v1/payment-methods', payload);
            } else if (isEdit && id) {
              await axios.put(`/api/v1/payment-methods/${id}`, payload);
            }
            navigate('/payment-methods');
          } catch (err: any) {
            setStatus({ success: false });
            setErrors({ submit: err?.error || err?.message || 'Request failed' });
            setSubmitting(false);
          }
        }}
      >
        {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, setFieldValue, touched, values }) => (
          <form onSubmit={handleSubmit}>
            <Stack spacing={2.5} sx={{ maxWidth: 560 }}>
              <TextField
                fullWidth
                label="Code"
                name="code"
                value={values.code}
                onBlur={handleBlur}
                onChange={handleChange}
                disabled={isView}
                error={Boolean(touched.code && errors.code)}
                helperText={(touched.code && errors.code) || 'Unique identifier, e.g. x402-base-usdc'}
              />

              {/* Common to both protocols. For x402 it's the network name (auto-filled on network select). */}
              <TextField
                fullWidth
                label="Name"
                name="name"
                value={values.name}
                onBlur={handleBlur}
                onChange={handleChange}
                disabled={isView}
                error={Boolean(touched.name && errors.name)}
                helperText={(touched.name && errors.name) || 'Human-readable name, e.g. Base Mainnet or Tempo Charge'}
              />

              <FormControl fullWidth error={Boolean(touched.protocol && errors.protocol)}>
                <InputLabel id="protocol-label">Protocol</InputLabel>
                <Select
                  labelId="protocol-label"
                  name="protocol"
                  value={values.protocol}
                  label="Protocol"
                  disabled={isView}
                  onChange={(e) => setFieldValue('protocol', e.target.value)}
                >
                  <MenuItem value="x402">x402</MenuItem>
                  <MenuItem value="mpp">MPP</MenuItem>
                </Select>
                {touched.protocol && errors.protocol && <FormHelperText>{errors.protocol}</FormHelperText>}
              </FormControl>

              {/* x402: pick a network (sets CAIP-2 + name) or enter it manually. */}
              {values.protocol === 'x402' && (
                <>
                  {!isView && (
                    <ToggleButtonGroup
                      exclusive
                      size="small"
                      value={networkMode}
                      onChange={(_, val) => {
                        if (val) setNetworkMode(val);
                      }}
                    >
                      <ToggleButton value="select">Select network</ToggleButton>
                      <ToggleButton value="custom">Custom</ToggleButton>
                    </ToggleButtonGroup>
                  )}

                  {!isView && networkMode === 'select' && (
                    <FormControl fullWidth>
                      <InputLabel id="network-label">Network</InputLabel>
                      <Select
                        labelId="network-label"
                        value={values.caip2_chain_id}
                        label="Network"
                        onChange={(e) => {
                          const selected = networks.find((n) => n.caip2 === e.target.value);
                          setFieldValue('caip2_chain_id', e.target.value);
                          setFieldValue('name', selected?.name ?? '');
                        }}
                      >
                        <MenuItem value="">
                          <em>None</em>
                        </MenuItem>
                        {networks.map((n) => (
                          <MenuItem key={n.caip2} value={n.caip2}>
                            {n.name}
                            <Typography component="span" variant="caption" color="text.secondary" sx={{ ml: 1 }}>
                              {n.caip2}
                            </Typography>
                          </MenuItem>
                        ))}
                      </Select>
                    </FormControl>
                  )}

                  {(isView || networkMode === 'custom') && (
                    <TextField
                      fullWidth
                      label="CAIP-2 Chain ID"
                      name="caip2_chain_id"
                      value={values.caip2_chain_id}
                      onBlur={handleBlur}
                      onChange={handleChange}
                      disabled={isView}
                      helperText="Optional, e.g. eip155:8453"
                    />
                  )}
                </>
              )}

              {/* MPP: no facilitator/network — a method (e.g. tempo) and scheme (charge). */}
              {values.protocol === 'mpp' && (
                <>
                  <FormControl fullWidth error={Boolean(touched.method && errors.method)}>
                    <InputLabel id="method-label">Method</InputLabel>
                    <Select
                      labelId="method-label"
                      name="method"
                      value={values.method}
                      label="Method"
                      disabled={isView}
                      onChange={(e) => setFieldValue('method', e.target.value)}
                    >
                      {METHODS.map((m) => (
                        <MenuItem key={m} value={m}>
                          {m}
                        </MenuItem>
                      ))}
                    </Select>
                    {touched.method && errors.method && <FormHelperText>{errors.method}</FormHelperText>}
                  </FormControl>
                  <FormControl fullWidth error={Boolean(touched.scheme && errors.scheme)}>
                    <InputLabel id="scheme-label">Scheme</InputLabel>
                    <Select
                      labelId="scheme-label"
                      name="scheme"
                      value={values.scheme}
                      label="Scheme"
                      disabled={isView}
                      onChange={(e) => setFieldValue('scheme', e.target.value)}
                    >
                      {SCHEMES.map((s) => (
                        <MenuItem key={s} value={s}>
                          {s}
                        </MenuItem>
                      ))}
                    </Select>
                    {touched.scheme && errors.scheme && <FormHelperText>{errors.scheme}</FormHelperText>}
                  </FormControl>
                </>
              )}

              <FormControlLabel
                control={
                  <Checkbox
                    name="enabled"
                    checked={values.enabled}
                    onChange={(e) => setFieldValue('enabled', e.target.checked)}
                    disabled={isView}
                    color="primary"
                  />
                }
                label="Enabled"
                sx={{ ml: 0 }}
              />

              <GlobalScopeToggle checked={values.is_global} onChange={(v) => setFieldValue('is_global', v)} disabled={isView} />

              {errors.submit && (
                <Box>
                  <FormHelperText error>{errors.submit as string}</FormHelperText>
                </Box>
              )}

              <Stack direction={{ xs: 'column-reverse', sm: 'row' }} spacing={2} justifyContent="flex-end">
                <Button variant="outlined" onClick={() => navigate('/payment-methods')} disabled={isSubmitting}>
                  {isView ? 'Back' : 'Cancel'}
                </Button>
                {isView && id && (
                  <Button
                    variant="contained"
                    startIcon={<EditTwoToneIcon />}
                    onClick={() => navigate('/payment-methods/edit', { state: { id } })}
                  >
                    Edit
                  </Button>
                )}
                {!isView && (
                  <Button type="submit" variant="contained" disabled={isSubmitting}>
                    {isCreate ? 'Create' : 'Save Changes'}
                  </Button>
                )}
              </Stack>
            </Stack>
          </form>
        )}
      </Formik>
    </MainCard>
  );
}
