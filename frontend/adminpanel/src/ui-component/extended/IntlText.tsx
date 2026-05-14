import { FormattedMessage } from 'react-intl';

interface IntlTextProps {
  id?: string;
  fallback?: string;
}

export default function IntlText({ id, fallback }: IntlTextProps) {
  const messageId = id ?? fallback;

  if (!messageId) {
    return null;
  }

  return <FormattedMessage id={messageId} defaultMessage={messageId} />;
}
