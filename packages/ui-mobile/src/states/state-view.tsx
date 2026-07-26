import {
  ActivityIndicator,
  StyleSheet,
  Text,
  View,
  type ViewProps,
} from "react-native";

import { Button } from "../button";
import { Card } from "../card";
import { useMobileTheme } from "../theme-provider";
import { mobileStateSemantics, type MobileStateKind } from "./model";

export interface MobileStateViewProps extends ViewProps {
  readonly kind: MobileStateKind;
  readonly title: string;
  readonly body: string;
  readonly actionLabel?: string;
  readonly onAction?: () => void;
}

export function MobileStateView({
  kind,
  title,
  body,
  actionLabel,
  onAction,
  style,
  ...props
}: MobileStateViewProps) {
  const theme = useMobileTheme();
  const semantics = mobileStateSemantics(kind);
  const showAction =
    semantics.actionAllowed && actionLabel !== undefined && onAction;

  return (
    <View
      accessibilityLiveRegion={semantics.liveRegion}
      accessibilityState={{ busy: semantics.busy }}
      style={[styles.container, style]}
      {...props}
    >
      <Card style={styles.card}>
        {kind === "loading" ? (
          <ActivityIndicator
            accessibilityLabel="Loading"
            color={theme.colors.accent}
            size="small"
          />
        ) : null}
        <Text
          accessibilityRole="header"
          style={[
            styles.title,
            {
              color: theme.colors.text,
              fontFamily: theme.typography.nativeFamilies.bold,
            },
          ]}
        >
          {title}
        </Text>
        <Text
          style={[
            styles.body,
            {
              color: theme.colors.textMuted,
              fontFamily: theme.typography.nativeFamilies.regular,
            },
          ]}
        >
          {body}
        </Text>
        {showAction ? <Button label={actionLabel} onPress={onAction} /> : null}
      </Card>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    alignItems: "stretch",
    width: "100%",
  },
  card: {
    gap: 12,
  },
  title: {
    fontSize: 20,
    lineHeight: 26,
  },
  body: {
    fontSize: 15,
    lineHeight: 22,
  },
});
