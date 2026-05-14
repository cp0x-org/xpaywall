import { createContext, ReactNode, useMemo } from 'react';

// project imports
import config from 'config';
import { useLocalStorage } from 'hooks/useLocalStorage';

// types
import { ConfigContextValue, ConfigStates } from 'types/config';

export type Props = {
  children: ReactNode;
};

// ==============================|| CONFIG CONTEXT ||============================== //

export const ConfigContext = createContext<ConfigContextValue | undefined>(undefined);

// ==============================|| CONFIG PROVIDER ||============================== //

export function ConfigProvider({ children }: Props) {
  const { state, setState, setField, resetState } = useLocalStorage<ConfigStates>('xpaywall-config', config);

  const memoizedValue = useMemo(() => ({ state, setState, setField, resetState }), [state, setField, setState, resetState]);

  return <ConfigContext.Provider value={memoizedValue}>{children}</ConfigContext.Provider>;
}
