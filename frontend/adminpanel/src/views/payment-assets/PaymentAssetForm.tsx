import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

// material-ui
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import CircularProgress from '@mui/material/CircularProgress';
import EditTwoToneIcon from '@mui/icons-material/EditTwoTone';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import axios from 'utils/axios';
import { PaymentMethodAsset } from './types';

interface PaymentMethodOption {
  id: string;
  name: string;
  code: string;
}

const emptyValues = {
  payment_method_id: '',
  symbol: '',
  contract_address: '',
  decimals: 18,
  submit: null
};

export default function PaymentAssetForm() {
  const { pathname, state } = useLocation();
  const navigate = useNavigate();

  const isCreate = pathname.includes('/create');
  const isEdit = pathname.includes('/edit');
  const isView = pathname.includes('/view');

  const id: string | undefined = (state as any)?.id;

  const [loading, setLoading] = useState(isEdit || isView);
  const [loadError, setLoadError] = useState('');
  const [initialValues, setInitialValues] = useState(emptyValues);
  const [paymentMethods, setPaymentMethods] = useState<PaymentMethodOption[]>([]);

  useEffect(() => {
    axios
      .get<PaymentMethodOption[]>('/api/v1/payment-methods')
      .then((res) => setPaymentMethods(res.data ?? []));
  }, []);

  useEffect(() => {
    if ((isEdit || isView) && id) {
      axios
        .get<PaymentMethodAsset>(`/api/v1/payment-method-assets/${id}`)
        .then((res) => {
          const d = res.data;
          setInitialValues({
            payment_method_id: d.payment_method_id,
            symbol: d.symbol,
            contract_address: d.contract_address ?? '',
            decimals: d.decimals,
            submit: null
          });
        })
        .catch(() => setLoadError('Failed to load payment asset'))
        .finally(() => setLoading(false));
    }
  }, [isEdit, isView, id]);

  let title = 'Payment Asset';
  if (isCreate) title = 'Create Payment Asset';
  if (isEdit) title = 'Update Payment Asset';
  if (isView) title = 'View Payment Asset';

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
          payment_method_id: Yup.string().required('Payment method is required'),
          symbol: Yup.string().required('Symbol is required'),
          decimals: Yup.number().integer().min(0).required('Decimals is required')
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          try {
            const payload = {
              payment_method_id: values.payment_method_id,
              symbol: values.symbol,
              contract_address: values.contract_address || null,
              decimals: Number(values.decimals)
            };
            if (isCreate) {
              await axios.post('/api/v1/payment-method-assets', payload);
            } else if (isEdit && id) {
              await axios.put(`/api/v1/payment-method-assets/${id}`, payload);
            }
            navigate('/payment-assets');
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
              <FormControl fullWidth error={Boolean(touched.payment_method_id && errors.payment_method_id)}>
                <InputLabel id="method-label">Payment Method</InputLabel>
                <Select
                  labelId="method-label"
                  name="payment_method_id"
                  value={values.payment_method_id}
                  label="Payment Method"
                  disabled={isView}
                  onChange={(e) => setFieldValue('payment_method_id', e.target.value)}
                >
                  {paymentMethods.map((m) => (
                    <MenuItem key={m.id} value={m.id}>
                      {m.name} ({m.code})
                    </MenuItem>
                  ))}
                </Select>
                {touched.payment_method_id && errors.payment_method_id && (
                  <FormHelperText>{errors.payment_method_id}</FormHelperText>
                )}
              </FormControl>

              <TextField
                fullWidth
                label="Symbol"
                name="symbol"
                value={values.symbol}
                onBlur={handleBlur}
                onChange={handleChange}
                disabled={isView}
                error={Boolean(touched.symbol && errors.symbol)}
                helperText={(touched.symbol && errors.symbol) || 'e.g. USDC'}
              />

              <TextField
                fullWidth
                label="Contract Address"
                name="contract_address"
                value={values.contract_address}
                onBlur={handleBlur}
                onChange={handleChange}
                disabled={isView}
                helperText="Optional for native tokens, e.g. 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
                slotProps={{ htmlInput: { style: { fontFamily: 'monospace' } } }}
              />

              <TextField
                fullWidth
                label="Decimals"
                name="decimals"
                type="number"
                value={values.decimals}
                onBlur={handleBlur}
                onChange={handleChange}
                disabled={isView}
                error={Boolean(touched.decimals && errors.decimals)}
                helperText={(touched.decimals && errors.decimals) || 'e.g. 6 for USDC, 18 for ETH'}
                slotProps={{ htmlInput: { min: 0, max: 77 } }}
              />

              {errors.submit && (
                <Box>
                  <FormHelperText error>{errors.submit as string}</FormHelperText>
                </Box>
              )}

              <Stack direction={{ xs: 'column-reverse', sm: 'row' }} spacing={2} justifyContent="flex-end">
                <Button variant="outlined" onClick={() => navigate('/payment-assets')} disabled={isSubmitting}>
                  {isView ? 'Back' : 'Cancel'}
                </Button>
                {isView && id && (
                  <Button
                    variant="contained"
                    startIcon={<EditTwoToneIcon />}
                    onClick={() => navigate('/payment-assets/edit', { state: { id } })}
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
