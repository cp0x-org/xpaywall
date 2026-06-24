import { useNavigate } from 'react-router-dom';

// material-ui
import Button from '@mui/material/Button';
import FormHelperText from '@mui/material/FormHelperText';
import InputLabel from '@mui/material/InputLabel';
import OutlinedInput from '@mui/material/OutlinedInput';
import Box from '@mui/material/Box';

// third party
import * as Yup from 'yup';
import { Formik } from 'formik';

// project imports
import AnimateButton from 'ui-component/extended/AnimateButton';
import CustomFormControl from 'ui-component/extended/Form/CustomFormControl';
import useAuth from 'hooks/useAuth';
import useScriptRef from 'hooks/useScriptRef';

import { useDispatch } from 'store';
import { openSnackbar } from 'store/slices/snackbar';

// extractToken pulls the `token` query param out of the reset URL returned by the API.
function extractToken(resetUrl: string): string | null {
  try {
    return new URL(resetUrl).searchParams.get('token');
  } catch {
    return null;
  }
}

// ========================|| JWT - FORGOT PASSWORD ||======================== //

export default function AuthForgotPassword({ ...others }) {
  const scriptedRef = useScriptRef();
  const dispatch = useDispatch();
  const navigate = useNavigate();

  const { requestPasswordReset } = useAuth();

  return (
    <Formik
      initialValues={{
        email: '',
        submit: null
      }}
      validationSchema={Yup.object().shape({
        email: Yup.string().email('Must be a valid email').max(255).required('Email is required')
      })}
      onSubmit={async (values, { setErrors, setStatus, setSubmitting }) => {
        try {
          // Backend always responds 200 (it never reveals whether the email exists).
          // Until SMTP is wired up it returns the reset link, which we use to route
          // straight to the reset page so the flow is usable end-to-end.
          const resetUrl = (await requestPasswordReset?.(values.email)) ?? '';
          const token = resetUrl ? extractToken(resetUrl) : null;

          if (scriptedRef.current) {
            setStatus({ success: true });
            setSubmitting(false);
            dispatch(
              openSnackbar({
                open: true,
                message: token ? 'Reset link generated. Redirecting…' : 'If the email is registered, a reset link has been sent.',
                variant: 'alert',
                alert: { color: 'success' },
                close: false
              })
            );
            setTimeout(() => {
              navigate(token ? `/reset-password?token=${encodeURIComponent(token)}` : '/login', { replace: true });
            }, 1500);
          }
        } catch (err: any) {
          console.error(err);
          if (scriptedRef.current) {
            setStatus({ success: false });
            setErrors({ submit: err?.error || err?.message || 'Request failed' });
            setSubmitting(false);
          }
        }
      }}
    >
      {({ errors, handleBlur, handleChange, handleSubmit, isSubmitting, touched, values }) => (
        <form noValidate onSubmit={handleSubmit} {...others}>
          <CustomFormControl fullWidth error={Boolean(touched.email && errors.email)}>
            <InputLabel htmlFor="outlined-adornment-email-forgot">Email Address</InputLabel>
            <OutlinedInput
              id="outlined-adornment-email-forgot"
              type="email"
              value={values.email}
              name="email"
              onBlur={handleBlur}
              onChange={handleChange}
              label="Email Address"
            />
            {touched.email && errors.email && (
              <FormHelperText error id="standard-weight-helper-text-email-forgot">
                {errors.email}
              </FormHelperText>
            )}
          </CustomFormControl>

          {errors.submit && (
            <Box sx={{ mt: 3 }}>
              <FormHelperText error>{errors.submit}</FormHelperText>
            </Box>
          )}

          <Box sx={{ mt: 2 }}>
            <AnimateButton>
              <Button disableElevation disabled={isSubmitting} fullWidth size="large" type="submit" variant="contained" color="secondary">
                Send Mail
              </Button>
            </AnimateButton>
          </Box>
        </form>
      )}
    </Formik>
  );
}
