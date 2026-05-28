import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

// material-ui
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import CircularProgress from '@mui/material/CircularProgress';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';
import Typography from '@mui/material/Typography';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import axios from 'utils/axios';

import { RouteRow } from './types';

interface Project {
  id: string;
  name: string;
  slug: string;
}

const emptyValues = {
  project_id: '',
  name: '',
  path_pattern: '',
  free: false,
  price_usd: '',
  description: '',
  submit: null
};

export default function RouteFormPage() {
  const { pathname, state } = useLocation();
  const navigate = useNavigate();

  const isCreate = pathname.includes('/create');
  const isEdit = pathname.includes('/edit');
  const isView = pathname.includes('/view');

  const routeId: string | undefined = (state as any)?.id;

  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(isEdit || isView);
  const [loadError, setLoadError] = useState('');
  const [initialValues, setInitialValues] = useState(emptyValues);

  let title = 'Route';
  if (isCreate) title = 'Create Route';
  if (isEdit) title = 'Update Route';
  if (isView) title = 'View Route';

  useEffect(() => {
    axios.get('/api/v1/projects').then((res) => setProjects(res.data ?? []));
  }, []);

  useEffect(() => {
    if ((isEdit || isView) && routeId) {
      axios
        .get<RouteRow>(`/api/v1/outbound-routes/${routeId}`)
        .then((res) => {
          const d = res.data;
          setInitialValues({
            project_id: d.project_id,
            name: d.name,
            path_pattern: d.path_pattern,
            free: d.free,
            price_usd: d.price_usd ?? '',
            description: d.description ?? '',
            submit: null
          });
        })
        .catch(() => setLoadError('Failed to load route data'))
        .finally(() => setLoading(false));
    }
  }, [isEdit, isView, routeId]);

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
          project_id: Yup.string().required('Project is required'),
          name: Yup.string().required('Name is required'),
          path_pattern: Yup.string().required('Path pattern is required'),
          price_usd: Yup.string().when('free', {
            is: false,
            then: (s) =>
              s
                .required('Price is required')
                .matches(/^\d+(\.\d+)?$/, 'Must be a positive number, e.g. 0.10')
          })
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          try {
            const payload = {
              project_id: values.project_id,
              name: values.name,
              path_pattern: values.path_pattern,
              free: values.free,
              price_amount: 0,
              price_usd: values.free ? '' : values.price_usd,
              description: values.description
            };
            if (isEdit && routeId) {
              await axios.put(`/api/v1/outbound-routes/${routeId}`, payload);
            } else {
              await axios.post('/api/v1/outbound-routes', payload);
            }
            navigate('/routes');
          } catch (err: any) {
            setStatus({ success: false });
            setErrors({ submit: err?.error || err?.message || 'Failed to save route' });
            setSubmitting(false);
          }
        }}
      >
        {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, setFieldValue, touched, values }) => (
          <form onSubmit={handleSubmit}>
            <Stack spacing={2} sx={{ maxWidth: 560 }}>
              <FormControl fullWidth error={Boolean(touched.project_id && errors.project_id)} disabled={isView}>
                <InputLabel id="project-label">Project</InputLabel>
                <Select
                  labelId="project-label"
                  name="project_id"
                  value={values.project_id}
                  label="Project"
                  onChange={(e) => setFieldValue('project_id', e.target.value)}
                >
                  {projects.map((p) => (
                    <MenuItem key={p.id} value={p.id}>
                      {p.name} ({p.slug})
                    </MenuItem>
                  ))}
                </Select>
                {touched.project_id && errors.project_id && (
                  <FormHelperText>{errors.project_id}</FormHelperText>
                )}
              </FormControl>

              <TextField
                fullWidth
                label="Route Name"
                name="name"
                value={values.name}
                onBlur={handleBlur}
                onChange={handleChange}
                error={Boolean(touched.name && errors.name)}
                helperText={touched.name && errors.name}
                disabled={isView}
              />

              <TextField
                fullWidth
                label="Path Pattern"
                name="path_pattern"
                value={values.path_pattern}
                onBlur={handleBlur}
                onChange={handleChange}
                error={Boolean(touched.path_pattern && errors.path_pattern)}
                helperText={(touched.path_pattern && errors.path_pattern) || 'e.g. /api/v1/* or /users/:id'}
                disabled={isView}
              />

              <TextField
                fullWidth
                multiline
                minRows={2}
                label="Description"
                name="description"
                value={values.description}
                onBlur={handleBlur}
                onChange={handleChange}
                disabled={isView}
              />

              <FormControlLabel
                control={
                  <Checkbox
                    name="free"
                    checked={values.free}
                    onChange={(e) => setFieldValue('free', e.target.checked)}
                    disabled={isView}
                  />
                }
                label="Free (no payment required)"
              />

              {!values.free && (
                <TextField
                  fullWidth
                  label="Price (USD)"
                  name="price_usd"
                  value={values.price_usd}
                  onBlur={handleBlur}
                  onChange={handleChange}
                  error={Boolean(touched.price_usd && errors.price_usd)}
                  helperText={(touched.price_usd && errors.price_usd) || 'e.g. 0.10'}
                  disabled={isView}
                />
              )}

              {errors.submit && (
                <Box>
                  <FormHelperText error>{errors.submit}</FormHelperText>
                </Box>
              )}

              <Stack direction="row" spacing={2} justifyContent="flex-end">
                <Button variant="outlined" onClick={() => navigate('/routes')} disabled={isSubmitting}>
                  {isView ? 'Back' : 'Cancel'}
                </Button>
                {!isView && (
                  <Button type="submit" variant="contained" disabled={isSubmitting}>
                    {isEdit ? 'Save Changes' : 'Create Route'}
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
