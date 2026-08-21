"use client";

import { Alert, Button, TextField } from "@mui/material";
import { useRouter } from "next/navigation";
import { useState } from "react";

export function AdminLogin() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [stage, setStage] = useState<"email" | "code">("email");
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");

  async function submit(action: "start" | "complete") {
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/auth", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action, email, password, code }),
      });
      const payload = (await response.json().catch(() => null)) as {
        message?: string;
      } | null;
      if (!response.ok)
        throw new Error(
          payload?.message || "Admin sign-in could not continue.",
        );
      if (action === "start") {
        setStage("code");
      } else {
        router.replace("/");
        router.refresh();
      }
    } catch (error) {
      setMessage(
        error instanceof Error
          ? error.message
          : "Admin sign-in could not continue.",
      );
    } finally {
      setBusy(false);
    }
  }

  return (
    <form
      className="admin-login-form"
      onSubmit={(event) => {
        event.preventDefault();
        void submit(stage === "email" ? "start" : "complete");
      }}
    >
      <TextField
        autoComplete="email"
        disabled={stage === "code"}
        fullWidth
        label="Admin email"
        onChange={(event) => setEmail(event.target.value)}
        required
        type="email"
        value={email}
      />
      <TextField
        autoComplete="current-password"
        disabled={stage === "code"}
        fullWidth
        label="Password"
        onChange={(event) => setPassword(event.target.value)}
        required
        type="password"
        value={password}
      />
      {stage === "code" ? (
        <TextField
          autoComplete="one-time-code"
          autoFocus
          fullWidth
          label="Six-digit code"
          onChange={(event) =>
            setCode(event.target.value.replace(/\D/g, "").slice(0, 6))
          }
          required
          slotProps={{
            htmlInput: {
              inputMode: "numeric",
              maxLength: 6,
              pattern: "[0-9]{6}",
            },
          }}
          value={code}
        />
      ) : null}
      {message ? <Alert severity="error">{message}</Alert> : null}
      <Button
        disabled={
          busy ||
          (stage === "email" && password.length === 0) ||
          (stage === "code" && code.length !== 6)
        }
        fullWidth
        type="submit"
        variant="contained"
      >
        {busy
          ? "Checking securely…"
          : stage === "email"
            ? "Send sign-in code"
            : "Verify and enter"}
      </Button>
      {stage === "code" ? (
        <Button
          disabled={busy}
          fullWidth
          onClick={() => {
            setStage("email");
            setCode("");
            setPassword("");
            setMessage("");
          }}
          type="button"
        >
          Use another email
        </Button>
      ) : null}
    </form>
  );
}
