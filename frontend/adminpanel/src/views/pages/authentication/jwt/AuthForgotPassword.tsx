import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

// material-ui
import Alert from '@mui/material/Alert';
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

// extractToken pulls the `token` query param out of the reset URL the API returns in dev (no SMTP).
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
  const navigate = useNavigate();
  const { requestPasswordReset } = useAuth();

  const [sentMessage, setSentMessage] = useState<string | null>(null);

  // After a successful request, replace the form with the backend's message.
  if (sentMessage) {
    return (
      <Alert severity="success" sx={{ mt: 1 }}>
        {sentMessage}
      </Alert>
    );
  }

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
          const { message, resetUrl } = await requestPasswordReset(values.email);
          const token = resetUrl ? extractToken(resetUrl) : null;

          if (scriptedRef.current) {
            setStatus({ success: true });
            setSubmitting(false);
            setSentMessage(message || 'If the email is registered, a reset link has been sent.');
            // Dev convenience: when SMTP is off the backend returns the link, so
            // jump straight to the reset page after the message is shown.
            if (token) {
              setTimeout(() => navigate(`/reset-password?token=${encodeURIComponent(token)}`, { replace: true }), 1500);
            }
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
