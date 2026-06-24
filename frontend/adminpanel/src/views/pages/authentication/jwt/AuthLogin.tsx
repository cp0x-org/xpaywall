import { MouseEvent, useState } from 'react';
import { Link } from 'react-router-dom';

// material-ui
import Button from '@mui/material/Button';
import Divider from '@mui/material/Divider';
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
import { useDispatch } from 'store';
import { openSnackbar } from 'store/slices/snackbar';

// assets
import Visibility from '@mui/icons-material/Visibility';
import VisibilityOff from '@mui/icons-material/VisibilityOff';

// ===============================|| JWT - LOGIN ||=============================== //

export default function JWTLogin({ ...others }) {
  const { login } = useAuth();
  const scriptedRef = useScriptRef();
  const dispatch = useDispatch();

  const [showPassword, setShowPassword] = useState(false);
  const handleClickShowPassword = () => {
    setShowPassword(!showPassword);
  };

  const handleMouseDownPassword = (event: MouseEvent) => {
    event.preventDefault()!;
  };

  return (
    <Formik
      initialValues={{
        username: '',
        password: '',
        submit: null
      }}
      validationSchema={Yup.object().shape({
        username: Yup.string().max(255).required('Username or email is required'),
        password: Yup.string().required('Password is required')
      })}
      onSubmit={async (values, { setStatus, setSubmitting }) => {
        try {
          await login?.(values.username, values.password);

          if (scriptedRef.current) {
            setStatus({ success: true });
            setSubmitting(false);
          }
        } catch (err: any) {
          console.error(err);
          const message = err.response?.data?.error || err.message || 'Login failed. Password or username incorrect.';
          dispatch(
            openSnackbar({
              open: true,
              message,
              variant: 'alert',
              alert: { variant: 'filled' },
              severity: 'error',
              anchorOrigin: { vertical: 'top', horizontal: 'center' },
              close: true
            })
          );
          if (scriptedRef.current) {
            setStatus({ success: false });
            setSubmitting(false);
          }
        }
      }}
    >
      {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values }) => (
        <form noValidate onSubmit={handleSubmit} {...others}>
          <CustomFormControl fullWidth error={Boolean(touched.username && errors.username)}>
            <InputLabel htmlFor="outlined-adornment-username-login">Username or Email</InputLabel>
            <OutlinedInput
              id="outlined-adornment-username-login"
              type="text"
              value={values.username}
              name="username"
              onBlur={handleBlur}
              onChange={handleChange}
              label="Username or Email"
            />
            {touched.username && errors.username && (
              <FormHelperText error id="standard-weight-helper-text-username-login">
                {errors.username}
              </FormHelperText>
            )}
          </CustomFormControl>

          <CustomFormControl fullWidth error={Boolean(touched.password && errors.password)}>
            <InputLabel htmlFor="outlined-adornment-password-login">Password</InputLabel>
            <OutlinedInput
              id="outlined-adornment-password-login"
              type={showPassword ? 'text' : 'password'}
              value={values.password}
              name="password"
              onBlur={handleBlur}
              onChange={handleChange}
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
              label="Password"
            />
            {touched.password && errors.password && (
              <FormHelperText error id="standard-weight-helper-text-password-login">
                {errors.password}
              </FormHelperText>
            )}
          </CustomFormControl>

          <Stack sx={{ alignItems: 'flex-end', mt: 1 }}>
            <Typography component={Link} to="/forgot-password" variant="subtitle1" sx={{ textDecoration: 'none' }} color="secondary">
              Forgot Password?
            </Typography>
          </Stack>

          <Box sx={{ mt: 2 }}>
            <AnimateButton>
              <Button color="secondary" disabled={isSubmitting} fullWidth size="large" type="submit" variant="contained">
                Sign In
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
  );
}
