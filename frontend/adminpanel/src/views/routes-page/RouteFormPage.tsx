import { useEffect, useState } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

// material-ui
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import InputLabel from '@mui/material/InputLabel';
import MenuItem from '@mui/material/MenuItem';
import Select from '@mui/material/Select';
import Stack from '@mui/material/Stack';
import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import FormControl from '@mui/material/FormControl';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import axios from 'utils/axios';

interface Project {
  id: string;
  name: string;
  slug: string;
}

export default function RouteFormPage() {
  const { pathname } = useLocation();
  const navigate = useNavigate();

  const [projects, setProjects] = useState<Project[]>([]);

  let title = 'Route';
  if (pathname.includes('/create')) title = 'Create Route';
  if (pathname.includes('/edit')) title = 'Update Route';
  if (pathname.includes('/view')) title = 'View Route';

  useEffect(() => {
    axios.get('/api/v1/projects').then((res) => setProjects(res.data ?? []));
  }, []);

  return (
    <MainCard title={title}>
      <Formik
        initialValues={{
          project_id: '',
          name: '',
          path_pattern: '',
          free: false,
          price_amount: 0,
          price_usd: '',
          description: '',
          submit: null
        }}
        validationSchema={Yup.object().shape({
          project_id: Yup.string().required('Project is required'),
          name: Yup.string().required('Name is required'),
          path_pattern: Yup.string().required('Path pattern is required'),
          price_amount: Yup.number().when('free', {
            is: false,
            then: (s) => s.min(0, 'Must be 0 or greater')
          })
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          try {
            await axios.post('/api/v1/outbound-routes', {
              project_id: values.project_id,
              name: values.name,
              path_pattern: values.path_pattern,
              free: values.free,
              price_amount: values.free ? 0 : Number(values.price_amount),
              price_usd: values.free ? '' : values.price_usd,
              description: values.description
            });
            navigate('/routes');
          } catch (err: any) {
            setStatus({ success: false });
            setErrors({ submit: err?.error || err?.message || 'Failed to create route' });
            setSubmitting(false);
          }
        }}
      >
        {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, setFieldValue, touched, values }) => (
          <form onSubmit={handleSubmit}>
            <Stack spacing={2} sx={{ maxWidth: 560 }}>
              <FormControl fullWidth error={Boolean(touched.project_id && errors.project_id)}>
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
              />

              <FormControlLabel
                control={
                  <Checkbox
                    name="free"
                    checked={values.free}
                    onChange={(e) => setFieldValue('free', e.target.checked)}
                  />
                }
                label="Free (no payment required)"
              />

              {!values.free && (
                <>
                  <TextField
                    fullWidth
                    label="Price Amount (cents)"
                    name="price_amount"
                    type="number"
                    value={values.price_amount}
                    onBlur={handleBlur}
                    onChange={handleChange}
                    error={Boolean(touched.price_amount && errors.price_amount)}
                    helperText={touched.price_amount && errors.price_amount}
                    slotProps={{ htmlInput: { min: 0 } }}
                  />

                  <TextField
                    fullWidth
                    label="Price USD"
                    name="price_usd"
                    value={values.price_usd}
                    onBlur={handleBlur}
                    onChange={handleChange}
                    helperText='e.g. "0.01"'
                  />
                </>
              )}

              {errors.submit && (
                <Box>
                  <FormHelperText error>{errors.submit}</FormHelperText>
                </Box>
              )}

              <Stack direction="row" spacing={2} justifyContent="flex-end">
                <Button variant="outlined" onClick={() => navigate('/routes')} disabled={isSubmitting}>
                  Cancel
                </Button>
                <Button type="submit" variant="contained" disabled={isSubmitting}>
                  Create Route
                </Button>
              </Stack>
            </Stack>
          </form>
        )}
      </Formik>
    </MainCard>
  );
}
