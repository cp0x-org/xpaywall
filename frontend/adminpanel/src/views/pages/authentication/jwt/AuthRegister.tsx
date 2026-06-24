import { SyntheticEvent, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useDispatch } from 'store';

// material-ui
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
import FormControl from '@mui/material/FormControl';
import FormHelperText from '@mui/material/FormHelperText';
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
import GoogleLoginButton from '../GoogleLoginButton';
import useAuth from 'hooks/useAuth';
import useScriptRef from 'hooks/useScriptRef';
import { strengthColor, strengthIndicator } from 'utils/password-strength';
import { openSnackbar } from 'store/slices/snackbar';
import { DASHBOARD_PATH } from 'config';

// assets
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';

// types
import { StringColorProps } from 'types';

// ===========================|| JWT - REGISTER ||=========================== //

export default function JWTRegister({ ...others }) {
  const navigate = useNavigate();
  const scriptedRef = useScriptRef();
  const dispatch = useDispatch();

  const [showPassword, setShowPassword] = useState(false);

  const [strength, setStrength] = useState(0);
  const [level, setLevel] = useState<StringColorProps>();
  const { register } = useAuth();

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
    changePassword('');
  }, []);

  return (
    <>
      <Stack sx={{ mb: 2, alignItems: 'center' }}>
        <Typography variant="subtitle1">Sign up with username & email</Typography>
      </Stack>

      <Formik
        initialValues={{
          username: '',
          email: '',
          password: '',
          submit: null
        }}
        validationSchema={Yup.object().shape({
          username: Yup.string()
            .trim()
            .required('Username is required')
            .min(2, 'Username must be at least 2 characters')
            .max(50, 'Username must not exceed 50 characters'),
          email: Yup.string().email('Must be a valid email').max(255).required('Email is required'),
          password: Yup.string()
            .required('Password is required')
            .min(8, 'Password must be at least 8 characters')
            .test('no-leading-trailing-whitespace', 'Password can not start or end with spaces', (value) => value === value?.trim())
        })}
        onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
          try {
            await register?.(values.username.trim(), values.email.trim(), values.password);
            if (scriptedRef.current) {
              setStatus({ success: true });
              setSubmitting(false);
              dispatch(
                openSnackbar({
                  open: true,
                  message: 'Your registration has been successfully completed.',
                  variant: 'alert',
                  alert: {
                    color: 'success'
                  },
                  close: false
                })
              );

              setTimeout(() => {
                navigate(DASHBOARD_PATH, { replace: true });
              }, 1000);
            }
          } catch (err: any) {
            console.error(err);
            if (scriptedRef.current) {
              setStatus({ success: false });
              setErrors({ submit: err?.error || err?.message || 'Registration failed' });
              setSubmitting(false);
            }
          }
        }}
      >
        {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values }) => (
          <form noValidate onSubmit={handleSubmit} {...others}>
            <CustomFormControl fullWidth error={Boolean(touched.username && errors.username)}>
              <InputLabel htmlFor="outlined-adornment-username-register">Username</InputLabel>
              <OutlinedInput
                id="outlined-adornment-username-register"
                type="text"
                value={values.username}
                name="username"
                onBlur={handleBlur}
                onChange={handleChange}
                label="Username"
              />
              {touched.username && errors.username && (
                <FormHelperText error id="standard-weight-helper-text-username-register">
                  {errors.username}
                </FormHelperText>
              )}
            </CustomFormControl>

            <CustomFormControl fullWidth error={Boolean(touched.email && errors.email)}>
              <InputLabel htmlFor="outlined-adornment-email-register">Email Address</InputLabel>
              <OutlinedInput
                id="outlined-adornment-email-register"
                type="email"
                value={values.email}
                name="email"
                onBlur={handleBlur}
                onChange={handleChange}
                label="Email Address"
              />
              {touched.email && errors.email && (
                <FormHelperText error id="standard-weight-helper-text-email-register">
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
                  <Stack direction="row" sx={{ gap: 2, alignItems: 'center' }}>
                    <Box sx={{ width: 85, height: 8, borderRadius: '7px', bgcolor: level?.color }} />
                    <Typography variant="subtitle1" sx={{ fontSize: '0.75rem' }}>
                      {level?.label}
                    </Typography>
                  </Stack>
                </Box>
              </FormControl>
            )}

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

            <Box sx={{ mt: 2 }}>
              <Divider>
                <Typography variant="caption" sx={{ color: 'text.secondary' }}>
                  OR
                </Typography>
              </Divider>
            </Box>

            <Box sx={{ mt: 2 }}>
              <GoogleLoginButton />
            </Box>
          </form>
        )}
      </Formik>
    </>
  );
}
