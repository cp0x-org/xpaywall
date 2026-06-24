import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

// material-ui
import Box from '@mui/material/Box';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import CircularProgress from '@mui/material/CircularProgress';
import EditTwoToneIcon from '@mui/icons-material/EditTwoTone';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Typography from '@mui/material/Typography';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import GlobalScopeToggle from 'ui-component/GlobalScopeToggle';
import useAuth from 'hooks/useAuth';
import axios from 'utils/axios';
import { Facilitator } from './types';

const emptyValues = {
  name: '',
  url: '',
  enabled: true,
  is_global: false,
  submit: null
};

export default function FacilitatorForm() {
  const { pathname, state } = useLocation();
  const navigate = useNavigate();

  const isCreate = pathname.includes('/create');
  const isEdit = pathname.includes('/edit');
  const isView = pathname.includes('/view');

  const id: string | undefined = (state as any)?.id;

  const { user } = useAuth();
  const isSuperadmin = user?.role === 'superadmin';

  const [loading, setLoading] = useState(isEdit || isView);
  const [loadError, setLoadError] = useState('');
  const [initialValues, setInitialValues] = useState(emptyValues);

  useEffect(() => {
    if ((isEdit || isView) && id) {
      axios
        .get<Facilitator>(`/api/v1/facilitators/${id}`)
        .then((res) => {
          const d = res.data;
          setInitialValues({
            name: d.name,
            url: d.url,
            enabled: d.enabled,
            is_global: d.is_global ?? false,
            submit: null
          });
        })
        .catch(() => setLoadError('Failed to load facilitator'))
        .finally(() => setLoading(false));
    }
  }, [isEdit, isView, id]);

  let title = 'Facilitator';
  if (isCreate) title = 'Create Facilitator';
  if (isEdit) title = 'Update Facilitator';
  if (isView) title = 'View Facilitator';

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
          name: Yup.string().required('Name is required'),
          url: Yup.string().url('Must be a valid URL').required('URL is required')
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          try {
            const payload = {
              name: values.name,
              url: values.url,
              enabled: values.enabled,
              is_global: values.is_global
            };
            if (isCreate) {
              await axios.post('/api/v1/facilitators', payload);
            } else if (isEdit && id) {
              await axios.put(`/api/v1/facilitators/${id}`, payload);
            }
            navigate('/facilitators');
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
                label="Name"
                name="name"
                value={values.name}
                onBlur={handleBlur}
                onChange={handleChange}
                disabled={isView}
                error={Boolean(touched.name && errors.name)}
                helperText={(touched.name && errors.name) || 'e.g. Coinbase x402 Facilitator'}
              />

              <TextField
                fullWidth
                label="URL"
                name="url"
                value={values.url}
                onBlur={handleBlur}
                onChange={handleChange}
                disabled={isView}
                error={Boolean(touched.url && errors.url)}
                helperText={(touched.url && errors.url) || 'e.g. https://facilitator.coinbase.com'}
              />

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
                <Button variant="outlined" onClick={() => navigate('/facilitators')} disabled={isSubmitting}>
                  {isView ? 'Back' : 'Cancel'}
                </Button>
                {/* Global entities are read-only for non-superadmins: hide Edit. */}
                {isView && id && (!initialValues.is_global || isSuperadmin) && (
                  <Button
                    variant="contained"
                    startIcon={<EditTwoToneIcon />}
                    onClick={() => navigate('/facilitators/edit', { state: { id } })}
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
