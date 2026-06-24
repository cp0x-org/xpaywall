// material-ui
import Box from '@mui/material/Box';
import Checkbox from '@mui/material/Checkbox';
import FormControlLabel from '@mui/material/FormControlLabel';
import FormHelperText from '@mui/material/FormHelperText';

// project imports
import useAuth from 'hooks/useAuth';

interface GlobalScopeToggleProps {
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
}

// GlobalScopeToggle lets a superadmin mark an entity (payment method / asset /
// facilitator) as global so every user can use it. It renders nothing for
// non-superadmins — the backend ignores is_global from them anyway.
export default function GlobalScopeToggle({ checked, onChange, disabled }: GlobalScopeToggleProps) {
  const { user } = useAuth();
  if (user?.role !== 'superadmin') return null;

  return (
    <Box>
      <FormControlLabel
        control={
          <Checkbox name="is_global" checked={checked} onChange={(e) => onChange(e.target.checked)} disabled={disabled} color="primary" />
        }
        label="Global (visible to all users)"
        sx={{ ml: 0 }}
      />
      <FormHelperText>Superadmin only — global entities can be used by every user.</FormHelperText>
    </Box>
  );
}
