"use client";

import {
  Alert,
  Button,
  IconButton,
  InputAdornment,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import { useRouter } from "next/navigation";
import { useState } from "react";

export function AdminLogin() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [stage, setStage] = useState<"email" | "code">("email");
  const [showPassword, setShowPassword] = useState(false);
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
      <div className="admin-login-progress" aria-label="Sign-in progress">
        <span className="is-active">1</span>
        <i aria-hidden="true" className={stage === "code" ? "is-active" : ""} />
        <span className={stage === "code" ? "is-active" : ""}>2</span>
        <Typography>
          {stage === "email" ? "Account details" : "Identity check"}
        </Typography>
      </div>
      {stage === "code" ? (
        <div className="admin-code-intro">
          <Typography component="h2">Check your inbox</Typography>
          <Typography>
            Enter the six-digit code sent to <strong>{email}</strong>.
          </Typography>
        </div>
      ) : null}
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
        slotProps={{
          input: {
            endAdornment: (
              <InputAdornment position="end">
                <Tooltip
                  arrow
                  title={showPassword ? "Hide password" : "Show password"}
                >
                  <IconButton
                    aria-label={
                      showPassword ? "Hide password" : "Show password"
                    }
                    disabled={stage === "code"}
                    edge="end"
                    onClick={() => setShowPassword((visible) => !visible)}
                    onMouseDown={(event) => event.preventDefault()}
                    type="button"
                  >
                    <span className="admin-password-icon" aria-hidden="true">
                      {showPassword ? "◉" : "◌"}
                    </span>
                  </IconButton>
                </Tooltip>
              </InputAdornment>
            ),
          },
        }}
        type={showPassword ? "text" : "password"}
        value={password}
      />
      {stage === "email" ? (
        <div className="admin-login-help">
          <a href="mailto:support@obiara.app?subject=Admin%20access%20recovery&body=I%20need%20help%20recovering%20access%20to%20the%20Obiara%20operations%20desk.%20Please%20send%20me%20the%20secure%20recovery%20steps.">
            Forgot password?
          </a>
        </div>
      ) : null}
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
        className="admin-auth-primary-action"
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
          className="admin-auth-secondary-action"
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
