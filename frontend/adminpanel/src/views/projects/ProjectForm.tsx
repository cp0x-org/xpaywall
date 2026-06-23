import * as React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';

// material-ui
import Button from '@mui/material/Button';
import EditTwoToneIcon from '@mui/icons-material/EditTwoTone';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import Grid from '@mui/material/Grid';
import Stack from '@mui/material/Stack';
import Tab from '@mui/material/Tab';
import Tabs from '@mui/material/Tabs';
import TextField from '@mui/material/TextField';
import Box from '@mui/material/Box';
import CircularProgress from '@mui/material/CircularProgress';
import Typography from '@mui/material/Typography';
import LinkTwoToneIcon from '@mui/icons-material/LinkTwoTone';
import LanguageTwoToneIcon from '@mui/icons-material/LanguageTwoTone';
import PersonTwoToneIcon from '@mui/icons-material/PersonTwoTone';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import MainCard from 'ui-component/cards/MainCard';
import useAuth from 'hooks/useAuth';
import axios from 'utils/axios';
import { canManage } from 'utils/ownership';

import { FullProject, ProxyUrl } from './types';
import { TableBody } from '../../ui-component/mui';
import ProjectPaymentMethods from './ProjectPaymentMethods';

function toSlug(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');
}

const emptyValues = {
  name: '',
  slug: '',
  owner_user_id: '',
  owner_username: '',
  base_url: '',
  auth_header_name: '',
  auth_header_value: '',
  allow_unmatched: false,
  submit: null
};

export default function ProjectFormPage() {
  const { pathname, state } = useLocation();
  const navigate = useNavigate();
  const { user } = useAuth();

  const isCreate = pathname.includes('/create');
  const isEdit = pathname.includes('/edit');
  const isView = pathname.includes('/view');

  const currentUserId = (user as { id?: string } | null | undefined)?.id;

  const projectId: string | undefined = (state as any)?.id;

  const [loading, setLoading] = React.useState(isEdit || isView);
  const [loadError, setLoadError] = React.useState('');
  const [initialValues, setInitialValues] = React.useState({
    ...emptyValues,
    owner_user_id: (user as any)?.id ?? '',
    owner_username: (user as any)?.username ?? ''
  });

  const [proxyUrl, setProxyUrl] = React.useState('');
  const [originTarget, setOriginTarget] = React.useState('');
  const [tab, setTab] = React.useState(0);

  React.useEffect(() => {
    axios
      .get<ProxyUrl>(`/api/v1/system/proxy-url`)
      .then((res) => {
        const d = res.data;
        setProxyUrl(d.proxy_url ?? '');
      })
      .catch(() => setLoadError('Failed to load proxy url'))
      .finally(() => setLoading(false));
  }, []);

  React.useEffect(() => {
    if ((isEdit || isView) && projectId) {
      axios
        .get<FullProject>(`/api/v1/projects/${projectId}/full`)
        .then((res) => {
          const d = res.data;
          setInitialValues({
            name: d.name,
            slug: d.slug,
            owner_user_id: d.owner_user_id,
            owner_username: d.owner_username ?? '',
            base_url: d.base_url ?? '',
            auth_header_name: d.auth_header_name ?? '',
            auth_header_value: d.auth_header_value ?? '',
            allow_unmatched: d.allow_unmatched ?? false,
            submit: null
          });
          setOriginTarget(d.base_url ?? '');
        })
        .catch(() => setLoadError('Failed to load project data'))
        .finally(() => setLoading(false));
    }
  }, [isEdit, isView, projectId]);

  let title = 'Project';
  if (isCreate) title = 'Create Project';
  if (isEdit) title = 'Update Project';
  if (isView) title = 'View Project';

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
      <Tabs value={tab} onChange={(_, v) => setTab(v)} sx={{ mb: 3, borderBottom: 1, borderColor: 'divider' }}>
        <Tab label="General" />
        <Tab label="Payment Methods" disabled={isCreate} />
      </Tabs>

      {tab === 1 && projectId && (
        <ProjectPaymentMethods
          projectId={projectId}
          isView={isView}
          canEdit={isCreate || canManage(currentUserId, initialValues.owner_user_id)}
        />
      )}

      {tab === 0 && (
        <>
          <Grid container spacing={3} sx={{ mb: 2 }}>
            <Grid size={{ xs: 12, sm: isEdit || isView ? 4 : 6 }}>
              <Grid container spacing={1} sx={{ alignItems: 'center' }}>
                <Grid>
                  <LinkTwoToneIcon color="primary" />
                </Grid>
                <Grid size={{ sm: 'grow' }}>
                  <Typography variant="h5" sx={{ wordBreak: 'break-all' }}>
                    {proxyUrl || '—'}
                  </Typography>
                  <Typography variant="subtitle2">PROXY URL</Typography>
                </Grid>
              </Grid>
            </Grid>
            <Grid size={{ xs: 12, sm: isEdit || isView ? 4 : 6 }}>
              <Grid container spacing={1} sx={{ alignItems: 'center' }}>
                <Grid>
                  <LanguageTwoToneIcon color="secondary" />
                </Grid>
                <Grid size={{ sm: 'grow' }}>
                  <Typography variant="h5" sx={{ wordBreak: 'break-all' }}>
                    {originTarget || '—'}
                  </Typography>
                  <Typography variant="subtitle2">ORIGIN TARGET</Typography>
                </Grid>
              </Grid>
            </Grid>
            {(isEdit || isView) && (
              <Grid size={{ xs: 12, sm: 4 }}>
                <Grid container spacing={1} sx={{ alignItems: 'center' }}>
                  <Grid>
                    <PersonTwoToneIcon color="action" />
                  </Grid>
                  <Grid size={{ sm: 'grow' }}>
                    <Typography variant="h5" sx={{ wordBreak: 'break-all' }}>
                      {initialValues.owner_username || '—'}
                    </Typography>
                    <Typography variant="subtitle2">OWNER</Typography>
                  </Grid>
                </Grid>
              </Grid>
            )}
          </Grid>
          <Box sx={{ mb: 3, width: '100%', height: 1, bgcolor: 'divider' }} />
          <Formik
            enableReinitialize
            initialValues={initialValues}
            validationSchema={Yup.object().shape({
              name: Yup.string().required('Name is required'),
              slug: Yup.string()
                .matches(/^[a-z0-9-]+$/, 'Only lowercase letters, numbers and hyphens')
                .required('Slug is required'),
              base_url: Yup.string().url('Must be a valid URL').required('Server Base URL is required')
            })}
            onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
              try {
                const payload = {
                  name: values.name,
                  slug: values.slug,
                  owner_user_id: values.owner_user_id,
                  base_url: values.base_url,
                  auth_header_name: values.auth_header_name || null,
                  auth_header_value: values.auth_header_value || null,
                  allow_unmatched: values.allow_unmatched
                };

                if (isCreate) {
                  await axios.post('/api/v1/projects', payload);
                } else if (isEdit && projectId) {
                  await axios.put(`/api/v1/projects/${projectId}`, payload);
                }

                navigate('/projects');
              } catch (err: any) {
                setStatus({ success: false });
                setErrors({ submit: err?.error || err?.message || 'Request failed' });
                setSubmitting(false);
              }
            }}
          >
            {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, setFieldValue, touched, values }) => (
              <form onSubmit={handleSubmit}>
                <Stack spacing={3} sx={{ width: '100%', maxWidth: 1080 }}>
                  <Box
                    sx={{
                      display: 'grid',
                      gridTemplateColumns: { xs: '1fr', lg: 'minmax(0, 1fr) minmax(0, 1fr)' },
                      gap: { xs: 3, lg: 4 },
                      alignItems: 'start'
                    }}
                  >
                    <Stack spacing={2.5}>
                      <Box>
                        <Typography variant="h4">Project Settings</Typography>
                        <Typography variant="body2" color="text.secondary">
                          Basic project identity and owner configuration.
                        </Typography>
                        <Box
                          sx={{
                            mt: 1.5,
                            width: '100%',
                            height: 1,
                            bgcolor: 'divider'
                          }}
                        />
                      </Box>

                      <TextField
                        fullWidth
                        label="Project Name"
                        name="name"
                        value={values.name}
                        onBlur={handleBlur}
                        disabled={isView}
                        onChange={(e) => {
                          handleChange(e);
                          if (!touched.slug) {
                            setFieldValue('slug', toSlug(e.target.value));
                          }
                        }}
                        error={Boolean(touched.name && errors.name)}
                        helperText={touched.name && errors.name}
                      />

                      <TextField
                        fullWidth
                        label="Slug"
                        name="slug"
                        value={values.slug}
                        onBlur={handleBlur}
                        disabled={isView}
                        onChange={handleChange}
                        error={Boolean(touched.slug && errors.slug)}
                        helperText={(touched.slug && errors.slug) || 'URL-friendly identifier, e.g. my-project'}
                      />
                    </Stack>

                    <Stack spacing={2.5}>
                      <Box>
                        <Typography variant="h4">Server Route Settings</Typography>
                        <Typography variant="body2" color="text.secondary">
                          Upstream server connection and route matching behavior.
                        </Typography>
                        <Box
                          sx={{
                            mt: 1.5,
                            width: '100%',
                            height: 1,
                            bgcolor: 'divider'
                          }}
                        />
                      </Box>

                      <TextField
                        fullWidth
                        label="Server Base URL"
                        name="base_url"
                        value={values.base_url}
                        onBlur={handleBlur}
                        disabled={isView}
                        onChange={(e) => {
                          handleChange(e);
                          setOriginTarget(e.target.value);
                        }}
                        error={Boolean(touched.base_url && errors.base_url)}
                        helperText={(touched.base_url && errors.base_url) || 'e.g. https://api.example.com'}
                      />

                      <TextField
                        fullWidth
                        label="Auth Header Name"
                        name="auth_header_name"
                        value={values.auth_header_name}
                        onBlur={handleBlur}
                        disabled={isView}
                        onChange={handleChange}
                        helperText="Optional, e.g. Authorization"
                      />

                      <TextField
                        fullWidth
                        label="Auth Header Value"
                        name="auth_header_value"
                        value={values.auth_header_value}
                        onBlur={handleBlur}
                        disabled={isView}
                        onChange={handleChange}
                        helperText="Optional, e.g. Bearer token123"
                      />

                      <FormControlLabel
                        control={
                          <Checkbox
                            name="allow_unmatched"
                            checked={values.allow_unmatched}
                            onChange={handleChange}
                            disabled={isView}
                            color="primary"
                          />
                        }
                        label="Allow Unmatched Routes"
                        sx={{ ml: 0 }}
                      />
                    </Stack>
                  </Box>

                  {errors.submit && (
                    <Box>
                      <FormHelperText error>{errors.submit as string}</FormHelperText>
                    </Box>
                  )}

                  <Stack direction={{ xs: 'column-reverse', sm: 'row' }} spacing={2} justifyContent="flex-end">
                    <Button variant="outlined" onClick={() => navigate('/projects')} disabled={isSubmitting}>
                      {isView ? 'Back' : 'Cancel'}
                    </Button>
                    {isView && projectId && (
                      <Button
                        variant="contained"
                        startIcon={<EditTwoToneIcon />}
                        onClick={() => navigate('/projects/edit', { state: { id: projectId } })}
                      >
                        Edit
                      </Button>
                    )}
                    {!isView && (isCreate || canManage(currentUserId, initialValues.owner_user_id)) && (
                      <Button type="submit" variant="contained" disabled={isSubmitting}>
                        {isCreate ? 'Create Project' : 'Save Changes'}
                      </Button>
                    )}
                  </Stack>
                </Stack>
              </form>
            )}
          </Formik>
        </>
      )}
    </MainCard>
  );
}
