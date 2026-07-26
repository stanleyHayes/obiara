import { useEffect, useMemo, useReducer, useRef } from "react";
import {
  PanResponder,
  StyleSheet,
  Text,
  View,
  type AccessibilityActionEvent,
} from "react-native";

import { Button } from "../button";
import { Pressable } from "../pressable";
import { useMobileTheme } from "../theme-provider";
import {
  HOLD_DURATION_MS,
  SOW_RELEASE_DISTANCE,
  initialGatherState,
  initialHoldState,
  initialSowState,
  initialStoneState,
  reduceGather,
  reduceHold,
  reduceSow,
  reduceStone,
} from "./model";

function useInterval(active: boolean, onTick: () => void) {
  const tick = useRef(onTick);
  tick.current = onTick;
  useEffect(() => {
    if (!active) return;
    const timer = setInterval(() => tick.current(), 50);
    return () => clearInterval(timer);
  }, [active]);
}

function Meter({ value }: Readonly<{ value: number }>) {
  const theme = useMobileTheme();
  return (
    <View
      accessibilityElementsHidden
      importantForAccessibility="no-hide-descendants"
      style={[styles.track, { backgroundColor: theme.colors.border }]}
    >
      <View
        style={[
          styles.fill,
          {
            backgroundColor: theme.colors.accent,
            width: `${Math.round(Math.min(1, Math.max(0, value)) * 100)}%`,
          },
        ]}
      />
    </View>
  );
}

function LabCard({
  eyebrow,
  title,
  consequence,
  children,
}: Readonly<{
  eyebrow: string;
  title: string;
  consequence: string;
  children: React.ReactNode;
}>) {
  const theme = useMobileTheme();
  return (
    <View
      style={[
        styles.card,
        {
          backgroundColor: theme.colors.surface,
          borderColor: theme.colors.border,
          borderRadius: theme.radii.large,
        },
      ]}
    >
      <Text style={[styles.eyebrow, { color: theme.colors.accent }]}>
        {eyebrow}
      </Text>
      <Text style={[styles.title, { color: theme.colors.text }]}>{title}</Text>
      <Text style={[styles.copy, { color: theme.colors.textMuted }]}>
        {consequence}
      </Text>
      {children}
    </View>
  );
}

export function HoldGesture() {
  const theme = useMobileTheme();
  const [state, dispatch] = useReducer(reduceHold, initialHoldState);
  useInterval(state.status === "holding", () =>
    dispatch({ type: "advance", milliseconds: 50 }),
  );
  const progress = state.elapsedMs / HOLD_DURATION_MS;
  const done = state.status === "completed";

  return (
    <LabCard
      consequence="Holding keeps this room private. Releasing early pauses safely; nothing is shared."
      eyebrow="HOLD · PRIVATE PAUSE"
      title={done ? "The room is held." : "Hold space before entering"}
    >
      <Meter value={progress} />
      <Pressable
        accessibilityActions={[
          { name: "activate", label: "Hold this room now" },
          { name: "escape", label: "Cancel holding" },
        ]}
        accessibilityHint="Press and hold for just over one second. Releasing early pauses without completing."
        accessibilityLabel={`Hold room, ${Math.round(progress * 100)} percent complete`}
        accessibilityRole="button"
        accessibilityState={{ disabled: done }}
        disabled={done}
        haptic="impact-light"
        onAccessibilityAction={(event: AccessibilityActionEvent) =>
          dispatch({
            type:
              event.nativeEvent.actionName === "escape" ? "cancel" : "confirm",
          })
        }
        onPressIn={() => dispatch({ type: "start" })}
        onPressOut={() => dispatch({ type: "release" })}
        style={[
          styles.gestureTarget,
          {
            backgroundColor: theme.colors.action,
            borderRadius: theme.radii.pill,
          },
        ]}
      >
        <Text style={[styles.targetLabel, { color: theme.colors.actionText }]}>
          {done
            ? "Held safely"
            : state.status === "paused"
              ? "Continue holding"
              : "Press and hold"}
        </Text>
      </Pressable>
      {!done && (
        <View style={styles.actionRow}>
          <Button
            label="Hold now"
            onPress={() => dispatch({ type: "confirm" })}
          />
          <Button
            label="Cancel"
            variant="secondary"
            onPress={() => dispatch({ type: "cancel" })}
          />
        </View>
      )}
    </LabCard>
  );
}

export function SowGesture() {
  const theme = useMobileTheme();
  const [state, dispatch] = useReducer(reduceSow, initialSowState);
  const pan = useMemo(
    () =>
      PanResponder.create({
        onMoveShouldSetPanResponder: (_, gesture) =>
          Math.abs(gesture.dx) > 8 || Math.abs(gesture.dy) > 8,
        onPanResponderMove: (_, gesture) =>
          dispatch({ type: "drag", distance: Math.max(0, -gesture.dy) }),
        onPanResponderRelease: () => dispatch({ type: "release" }),
        onPanResponderTerminate: () => dispatch({ type: "release" }),
      }),
    [],
  );
  const staged = state.status !== "recording" && state.status !== "sown";
  const progress = state.distance / SOW_RELEASE_DISTANCE;

  return (
    <LabCard
      consequence="Sowing publishes this voice seed to your circle. You can review it before release."
      eyebrow="SOW · DELIBERATE SHARE"
      title={
        state.status === "sown"
          ? "Your seed is sown."
          : "Stage before you share"
      }
    >
      {state.status === "recording" ? (
        <Button
          accessibilityHint="Stops the recording and opens a review step"
          label="Finish recording"
          onPress={() => dispatch({ type: "finish-recording" })}
        />
      ) : (
        <>
          <Meter value={progress} />
          <Pressable
            {...pan.panHandlers}
            accessibilityActions={[
              { name: "activate", label: "Sow this seed" },
              { name: "escape", label: "Discard and record again" },
            ]}
            accessibilityHint="Drag upward and release, or use the visible Sow seed button"
            accessibilityLabel={
              state.status === "sown"
                ? "Voice seed sown"
                : `Staged voice seed, ${Math.round(progress * 100)} percent toward release`
            }
            accessibilityRole="button"
            onAccessibilityAction={(event: AccessibilityActionEvent) =>
              dispatch({
                type:
                  event.nativeEvent.actionName === "escape"
                    ? "reset"
                    : "confirm",
              })
            }
            style={[
              styles.gestureTarget,
              {
                backgroundColor: theme.colors.accent,
                borderRadius: theme.radii.large,
              },
            ]}
          >
            <Text
              style={[styles.targetLabel, { color: theme.colors.accentText }]}
            >
              {state.status === "sown"
                ? "Seed released"
                : state.status === "ready"
                  ? "Release to sow"
                  : "Drag up to sow"}
            </Text>
          </Pressable>
          {staged && (
            <View style={styles.actionRow}>
              <Button
                label="Sow seed"
                onPress={() => dispatch({ type: "confirm" })}
              />
              <Button
                label="Record again"
                variant="secondary"
                onPress={() => dispatch({ type: "reset" })}
              />
            </View>
          )}
        </>
      )}
    </LabCard>
  );
}

export function StoneGesture() {
  const theme = useMobileTheme();
  const [state, dispatch] = useReducer(reduceStone, initialStoneState);
  useInterval(state.status === "settling", () =>
    dispatch({ type: "advance", milliseconds: 50 }),
  );
  const pan = useMemo(
    () =>
      PanResponder.create({
        onMoveShouldSetPanResponder: (_, gesture) => gesture.dy > 8,
        onPanResponderMove: (_, gesture) =>
          dispatch({ type: "drag", distance: gesture.dy }),
        onPanResponderRelease: () => dispatch({ type: "release" }),
        onPanResponderTerminate: () => dispatch({ type: "release" }),
      }),
    [],
  );

  return (
    <LabCard
      consequence="Setting the stone closes this turn and saves the reflection. It does not notify anyone."
      eyebrow="STONE · CLOSE A TURN"
      title={
        state.status === "settled"
          ? "The stone is settled."
          : "Let the thought settle"
      }
    >
      <Meter value={state.progress} />
      <Pressable
        {...pan.panHandlers}
        accessibilityActions={[
          { name: "activate", label: "Settle this reflection" },
          { name: "escape", label: "Reset the stone" },
        ]}
        accessibilityHint="Press and hold or move slowly downward. The Settle now button does the same thing."
        accessibilityLabel={`Settle reflection, ${Math.round(state.progress * 100)} percent complete`}
        accessibilityRole="button"
        haptic="impact-light"
        onAccessibilityAction={(event: AccessibilityActionEvent) =>
          dispatch({
            type:
              event.nativeEvent.actionName === "escape" ? "reset" : "confirm",
          })
        }
        onPressIn={() => dispatch({ type: "start" })}
        onPressOut={() => dispatch({ type: "release" })}
        style={[
          styles.stone,
          {
            backgroundColor: theme.colors.text,
            borderRadius: theme.radii.pill,
            transform: theme.motion.reduceMotion
              ? undefined
              : [{ translateY: state.progress * 20 }],
          },
        ]}
      >
        <Text style={[styles.targetLabel, { color: theme.colors.surface }]}>
          {state.status === "settled" ? "Settled" : "Hold or draw down"}
        </Text>
      </Pressable>
      {state.status !== "settled" && (
        <Button
          label="Settle now"
          onPress={() => dispatch({ type: "confirm" })}
        />
      )}
    </LabCard>
  );
}

export function GatherGesture() {
  const theme = useMobileTheme();
  const [state, dispatch] = useReducer(reduceGather, initialGatherState);
  const startDistance = useRef<number | undefined>(undefined);
  const pan = useMemo(
    () =>
      PanResponder.create({
        onStartShouldSetPanResponder: (event) =>
          event.nativeEvent.touches.length === 2,
        onMoveShouldSetPanResponder: (event) =>
          event.nativeEvent.touches.length === 2,
        onPanResponderGrant: (event) => {
          const [first, second] = event.nativeEvent.touches;
          if (first && second) {
            startDistance.current = Math.hypot(
              first.pageX - second.pageX,
              first.pageY - second.pageY,
            );
          }
        },
        onPanResponderMove: (event) => {
          const [first, second] = event.nativeEvent.touches;
          if (!first || !second || !startDistance.current) return;
          const distance = Math.hypot(
            first.pageX - second.pageX,
            first.pageY - second.pageY,
          );
          dispatch({
            type: "set",
            amount: 0.5 + (distance - startDistance.current) / 200,
          });
        },
        onPanResponderRelease: () => {
          startDistance.current = undefined;
        },
      }),
    [],
  );
  const percent = Math.round(state.amount * 100);

  return (
    <LabCard
      consequence="Gathering sets how much of the circle appears together. Confirming only changes this view."
      eyebrow="GATHER · SHAPE THE CIRCLE"
      title={
        state.completed ? `${percent}% gathered.` : "Bring the circle closer"
      }
    >
      <View
        {...pan.panHandlers}
        accessibilityActions={[
          { name: "increment", label: "Gather ten percent more" },
          { name: "decrement", label: "Gather ten percent less" },
          { name: "activate", label: "Confirm this gathering" },
        ]}
        accessibilityHint="Pinch with two fingers, use accessibility adjust actions, or use the visible minus and plus buttons"
        accessibilityLabel="Gather amount"
        accessibilityRole="adjustable"
        accessibilityValue={{
          min: 0,
          max: 100,
          now: percent,
          text: `${percent} percent`,
        }}
        onAccessibilityAction={(event: AccessibilityActionEvent) => {
          if (event.nativeEvent.actionName === "increment") {
            dispatch({ type: "adjust", delta: 0.1 });
          } else if (event.nativeEvent.actionName === "decrement") {
            dispatch({ type: "adjust", delta: -0.1 });
          } else {
            dispatch({ type: "confirm" });
          }
        }}
        style={[styles.gatherSurface, { borderColor: theme.colors.border }]}
      >
        <Text style={[styles.gatherValue, { color: theme.colors.text }]}>
          {percent}%
        </Text>
        <Meter value={state.amount} />
      </View>
      <View style={styles.actionRow}>
        <Button
          accessibilityLabel="Gather less"
          label="−"
          variant="secondary"
          onPress={() => dispatch({ type: "adjust", delta: -0.1 })}
        />
        <Button
          accessibilityLabel="Gather more"
          label="+"
          variant="secondary"
          onPress={() => dispatch({ type: "adjust", delta: 0.1 })}
        />
      </View>
      <Button
        label="Confirm gathering"
        onPress={() => dispatch({ type: "confirm" })}
      />
    </LabCard>
  );
}

const styles = StyleSheet.create({
  actionRow: { flexDirection: "row", flexWrap: "wrap", gap: 12 },
  card: { borderWidth: 1, gap: 16, padding: 20 },
  copy: { fontFamily: "Outfit_400Regular", fontSize: 16, lineHeight: 24 },
  eyebrow: { fontFamily: "Outfit_700Bold", fontSize: 12, letterSpacing: 1.2 },
  fill: { borderRadius: 999, height: "100%" },
  gatherSurface: { borderWidth: 1, gap: 12, minHeight: 120, padding: 20 },
  gatherValue: { fontFamily: "Outfit_700Bold", fontSize: 32 },
  gestureTarget: {
    alignItems: "center",
    justifyContent: "center",
    minHeight: 64,
    paddingHorizontal: 20,
    paddingVertical: 14,
  },
  stone: {
    alignItems: "center",
    alignSelf: "center",
    height: 112,
    justifyContent: "center",
    padding: 12,
    width: 112,
  },
  targetLabel: {
    fontFamily: "Outfit_700Bold",
    fontSize: 16,
    textAlign: "center",
  },
  title: { fontFamily: "Outfit_700Bold", fontSize: 25, lineHeight: 31 },
  track: { borderRadius: 999, height: 8, overflow: "hidden", width: "100%" },
});
