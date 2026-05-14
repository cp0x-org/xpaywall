import { SyntheticEvent, useEffect, useState } from 'react';
import { Link } from 'react-router-dom';

// material-ui
import { useTheme } from '@mui/material/styles';
import Button from '@mui/material/Button';
import Checkbox from '@mui/material/Checkbox';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';
import Grid from '@mui/material/Grid';
import IconButton from '@mui/material/IconButton';
import InputAdornment from '@mui/material/InputAdornment';
import InputLabel from '@mui/material/InputLabel';
import OutlinedInput from '@mui/material/OutlinedInput';
import Stack from '@mui/material/Stack';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import AnimateButton from 'ui-component/extended/AnimateButton';
import CustomFormControl from 'ui-component/extended/Form/CustomFormControl';

import useAuth from 'hooks/useAuth';
import useConfig from 'hooks/useConfig';
import useScriptRef from 'hooks/useScriptRef';
import { withAlpha } from 'utils/colorUtils';
import { strengthColor, strengthIndicator } from 'utils/password-strength';

// types
import { StringColorProps } from 'types';

// assets
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';
import FirebaseSocial from './FirebaseSocial';

// ===========================|| FIREBASE - REGISTER ||=========================== //

export default function FirebaseRegister({ ...others }) {
  const theme = useTheme();
  const scriptedRef = useScriptRef();
  const [showPassword, setShowPassword] = useState(false);
  const [checked, setChecked] = useState(true);

  const {
    state: { borderRadius }
  } = useConfig();

  const [strength, setStrength] = useState(0);
  const [level, setLevel] = useState<StringColorProps>();
  const { firebaseRegister } = useAuth();

  const handleClickShowPassword = () => {
    setShowPassword(!showPassword);
  };

  const handleMouseDownPassword = (event: SyntheticEvent) => {
    event.preventDefault();
  };

  const changePassword = (value: string) => {
    const temp = strengthIndicator(value);
    setStrength(temp);
    setLevel(strengthColor(temp));
  };

  useEffect(() => {
    changePassword('123456');
  }, []);

  return (
    <>
      <Stack sx={{ gap: 2 }}>
        <FirebaseSocial />
        <Box sx={{ alignItems: 'center', display: 'flex' }}>
          <Divider sx={{ flexGrow: 1 }} orientation="horizontal" />
          <Button
            variant="outlined"
            sx={{
              cursor: 'unset',
              m: 2,
              py: 0.5,
              px: 7,
              borderColor: `${theme.vars.palette.grey[100]} !important`,
              color: `${theme.vars.palette.grey[900]}!important`,
              fontWeight: 500,
              borderRadius: `${borderRadius}px`,
              ...theme.applyStyles('dark', {
                borderColor: `${withAlpha(theme.vars.palette.dark.light, 0.2)} !important`
              })
            }}
            disableRipple
            disabled
          >
            OR
          </Button>
          <Divider sx={{ flexGrow: 1 }} orientation="horizontal" />
        </Box>
        <Stack sx={{ alignItems: 'center', mb: 2 }}>
          <Typography variant="subtitle1">Sign up with Email address</Typography>
        </Stack>
      </Stack>
      <Formik
        initialValues={{
          email: '',
          password: '',
          submit: null
        }}
        validationSchema={Yup.object().shape({
          email: Yup.string().email('Must be a valid email').max(255).required('Email is required'),
          password: Yup.string().max(255).required('Password is required')
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          try {
            await firebaseRegister?.(values.email, values.password);
            if (scriptedRef.current) {
              setStatus({ success: true });
              setSubmitting(false);
            }
          } catch (err: any) {
            console.error(err);
            if (scriptedRef.current) {
              setStatus({ success: false });
              setErrors({ submit: err.message });
              setSubmitting(false);
            }
          }
        }}
      >
        {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values }) => (
          <form noValidate onSubmit={handleSubmit} {...others}>
            <Grid container spacing={{ xs: 0, md: 2 }}>
              <Grid size={{ xs: 12, sm: 6 }}>
                <CustomFormControl fullWidth>
                  <InputLabel htmlFor="outlined-adornment-first-register">First Name</InputLabel>
                  <OutlinedInput id="outlined-adornment-first-register" type="text" name="firstName" />
                </CustomFormControl>
              </Grid>
              <Grid size={{ xs: 12, sm: 6 }}>
                <CustomFormControl fullWidth>
                  <InputLabel htmlFor="outlined-adornment-last-register">Last Name</InputLabel>
                  <OutlinedInput id="outlined-adornment-last-register" type="text" name="lastName" />
                </CustomFormControl>
              </Grid>
            </Grid>
            <CustomFormControl fullWidth error={Boolean(touched.email && errors.email)}>
              <InputLabel htmlFor="outlined-adornment-email-register">Email Address / Username</InputLabel>
              <OutlinedInput
                id="outlined-adornment-email-register"
                type="email"
                value={values.email}
                name="email"
                onBlur={handleBlur}
                onChange={handleChange}
              />
              {touched.email && errors.email && (
                <FormHelperText error id="standard-weight-helper-text--register">
                  {errors.email}
                </FormHelperText>
              )}
            </CustomFormControl>

            <CustomFormControl fullWidth error={Boolean(touched.password && errors.password)}>
              <InputLabel htmlFor="outlined-adornment-password-register">Password</InputLabel>
              <OutlinedInput
                id="outlined-adornment-password-register"
                type={showPassword ? 'text' : 'password'}
                value={values.password}
                name="password"
                label="Password"
                onBlur={handleBlur}
                onChange={(e) => {
                  handleChange(e);
                  changePassword(e.target.value);
                }}
                endAdornment={
                  <InputAdornment position="end">
                    <IconButton
                      aria-label="toggle password visibility"
                      onClick={handleClickShowPassword}
                      onMouseDown={handleMouseDownPassword}
                      edge="end"
                      size="large"
                    >
                      {showPassword ? <Visibility /> : <VisibilityOff />}
                    </IconButton>
                  </InputAdornment>
                }
              />
              {touched.password && errors.password && (
                <FormHelperText error id="standard-weight-helper-text-password-register">
                  {errors.password}
                </FormHelperText>
              )}
            </CustomFormControl>

            {strength !== 0 && (
              <FormControl fullWidth>
                <Box sx={{ mb: 2 }}>
                  <Stack direction="row" sx={{ alignItems: 'center', gap: 2 }}>
                    <Box sx={{ width: 85, height: 8, borderRadius: '7px', bgcolor: level?.color }} />
                    <Typography variant="subtitle1" sx={{ fontSize: '0.75rem' }}>
                      {level?.label}
                    </Typography>
                  </Stack>
                </Box>
              </FormControl>
            )}

            <FormControlLabel
              control={<Checkbox checked={checked} onChange={(event) => setChecked(event.target.checked)} name="checked" color="primary" />}
              label={
                <Typography variant="subtitle1">
                  Agree with &nbsp;
                  <Typography variant="subtitle1" component={Link} to="#">
                    Terms & Condition.
                  </Typography>
                </Typography>
              }
            />
            {errors.submit && (
              <Box sx={{ mt: 3 }}>
                <FormHelperText error>{errors.submit}</FormHelperText>
              </Box>
            )}

            <Box sx={{ mt: 2 }}>
              <AnimateButton>
                <Button disableElevation disabled={isSubmitting} fullWidth size="large" type="submit" variant="contained" color="secondary">
                  Sign up
                </Button>
              </AnimateButton>
            </Box>
          </form>
        )}
      </Formik>
    </>
  );
}
