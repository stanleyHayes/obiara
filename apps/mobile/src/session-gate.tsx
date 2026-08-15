import { useEffect, useState, type ReactNode } from "react";
import {
  ActivityIndicator,
  Image,
  KeyboardAvoidingView,
  Platform,
  Pressable,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";

import brandMark from "../assets/brand-mark.png";
import { accessToken, apiRequest, onSessionCleared, verifyOtp } from "./api";

type Stage = "checking" | "phone" | "code" | "signed-in";

export function SessionGate({ children }: Readonly<{ children: ReactNode }>) {
  const [stage, setStage] = useState<Stage>("checking");
  const [phone, setPhone] = useState("+233");
  const [code, setCode] = useState("");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  useEffect(() => {
    let active = true;
    void accessToken().then((token) => {
      if (active) setStage(token ? "signed-in" : "phone");
    });
    const unsubscribe = onSessionCleared((expired) => {
      setStage("phone");
      setCode("");
      setMessage(
        expired ? "Your sign-in has expired. Please sign in again." : "",
      );
    });
    return () => {
      active = false;
      unsubscribe();
    };
  }, []);

  async function requestCode() {
    setBusy(true);
    setMessage("");
    try {
      await apiRequest<{ status: string }>(
        "/v1/auth/otp",
        { method: "POST", body: JSON.stringify({ phone: phone.trim() }) },
        false,
      );
      setStage("code");
    } catch (error) {
      setMessage(
        error instanceof Error ? error.message : "The code could not be sent.",
      );
    } finally {
      setBusy(false);
    }
  }

  async function verifyCode() {
    setBusy(true);
    setMessage("");
    try {
      await verifyOtp(phone.trim(), code);
      setStage("signed-in");
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "The code could not be verified.",
      );
    } finally {
      setBusy(false);
    }
  }

  if (stage === "signed-in") return children;
  if (stage === "checking") {
    return (
      <View style={styles.loading}>
        <ActivityIndicator color="#8E3159" />
      </View>
    );
  }

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === "ios" ? "padding" : undefined}
      style={styles.page}
    >
      <View style={styles.card}>
        <Image
          accessibilityIgnoresInvertColors
          source={brandMark}
          style={styles.mark}
        />
        <Text style={styles.kicker}>YOUR PRIVATE COMPOUND</Text>
        <Text style={styles.title}>
          {stage === "phone" ? "Come home to Obiara." : "One last step."}
        </Text>
        <Text style={styles.copy}>
          {stage === "phone"
            ? "Use the phone number tied to your membership. We will send a short-lived code."
            : `Enter the six-digit code sent to ${phone}.`}
        </Text>
        <TextInput
          accessibilityLabel={
            stage === "phone" ? "Phone number" : "Six-digit code"
          }
          autoComplete={stage === "phone" ? "tel" : "one-time-code"}
          editable={!busy && stage === "phone"}
          keyboardType="phone-pad"
          maxLength={stage === "phone" ? 16 : 6}
          onChangeText={(value) =>
            stage === "phone"
              ? setPhone(value)
              : setCode(value.replace(/\D/g, "").slice(0, 6))
          }
          placeholder={stage === "phone" ? "+233 24 000 0000" : "000000"}
          style={styles.input}
          value={stage === "phone" ? phone : code}
        />
        {message ? (
          <Text accessibilityLiveRegion="polite" style={styles.error}>
            {message}
          </Text>
        ) : null}
        <Pressable
          accessibilityRole="button"
          disabled={busy || (stage === "code" && code.length !== 6)}
          onPress={() =>
            void (stage === "phone" ? requestCode() : verifyCode())
          }
          style={({ pressed }) => [
            styles.button,
            (pressed || busy) && styles.buttonPressed,
          ]}
        >
          <Text style={styles.buttonText}>
            {busy
              ? "Checking securely…"
              : stage === "phone"
                ? "Send sign-in code"
                : "Verify and enter"}
          </Text>
        </Pressable>
        {stage === "code" ? (
          <Pressable
            accessibilityRole="button"
            disabled={busy}
            onPress={() => {
              setStage("phone");
              setCode("");
              setMessage("");
            }}
          >
            <Text style={styles.link}>Use another number</Text>
          </Pressable>
        ) : null}
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  loading: {
    alignItems: "center",
    backgroundColor: "#FFF3E6",
    flex: 1,
    justifyContent: "center",
  },
  page: {
    alignItems: "center",
    backgroundColor: "#F3E8DF",
    flex: 1,
    justifyContent: "center",
    padding: 24,
  },
  card: {
    backgroundColor: "#FFFDFC",
    borderColor: "rgba(58,14,46,0.12)",
    borderRadius: 18,
    borderWidth: 1,
    maxWidth: 420,
    padding: 28,
    width: "100%",
  },
  mark: { height: 48, marginBottom: 28, resizeMode: "contain", width: 48 },
  kicker: {
    color: "#8E3159",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 11,
    letterSpacing: 1.6,
  },
  title: {
    color: "#26101F",
    fontFamily: "Outfit_800ExtraBold",
    fontSize: 36,
    letterSpacing: -1.5,
    lineHeight: 38,
    marginTop: 10,
  },
  copy: {
    color: "#765F70",
    fontFamily: "Outfit_400Regular",
    fontSize: 16,
    lineHeight: 24,
    marginBottom: 24,
    marginTop: 14,
  },
  input: {
    backgroundColor: "#FFFFFF",
    borderColor: "#D8C6CD",
    borderRadius: 12,
    borderWidth: 1,
    color: "#26101F",
    fontFamily: "Outfit_600SemiBold",
    fontSize: 18,
    minHeight: 54,
    paddingHorizontal: 16,
  },
  error: { color: "#9D2948", fontFamily: "Outfit_600SemiBold", marginTop: 12 },
  button: {
    alignItems: "center",
    backgroundColor: "#3A0E2E",
    borderRadius: 12,
    justifyContent: "center",
    marginTop: 16,
    minHeight: 54,
    paddingHorizontal: 18,
  },
  buttonPressed: { opacity: 0.72 },
  buttonText: { color: "#FFF3E6", fontFamily: "Outfit_700Bold", fontSize: 16 },
  link: {
    color: "#8E3159",
    fontFamily: "Outfit_700Bold",
    marginTop: 18,
    textAlign: "center",
  },
});
