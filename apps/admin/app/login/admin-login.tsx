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
import { SegmentedOtpInput } from "@obiara/ui-web";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { isAdminSessionResult, isCodeSent } from "../auth-model";
import { AdminSkeleton } from "../loading-skeleton";

export function AdminLogin() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [code, setCode] = useState("");
  const [stage, setStage] = useState<"email" | "code">("email");
  const [showPassword, setShowPassword] = useState(false);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState("");
  const mounted = useRef(false);
  const generation = useRef(0);
  const controller = useRef<AbortController | null>(null);
  const emailField = useRef<HTMLInputElement>(null);

  useEffect(() => {
    mounted.current = true;
    const lifecycle = generation;
    return () => {
      mounted.current = false;
      lifecycle.current++;
      controller.current?.abort();
    };
  }, []);

  async function submit(action: "start" | "complete") {
    const run = ++generation.current;
    controller.current?.abort();
    const requestController = new AbortController();
    controller.current = requestController;
    setBusy(true);
    setMessage("");
    try {
      const response = await fetch("/api/auth", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(
          action === "start"
            ? { action, email, password }
            : { action, email, code },
        ),
        signal: requestController.signal,
      });
      const payload: unknown = await response.json().catch(() => null);
      if (!response.ok)
        throw new Error(
          payload &&
            typeof payload === "object" &&
            "message" in payload &&
            typeof payload.message === "string"
            ? payload.message
            : "Admin sign-in could not continue.",
        );
      if (action === "start") {
        if (!isCodeSent(payload))
          throw new Error("Admin sign-in could not continue.");
        if (!mounted.current || run !== generation.current) return;
        setPassword("");
        setShowPassword(false);
        setStage("code");
      } else {
        if (!isAdminSessionResult(payload))
          throw new Error("Admin sign-in could not continue.");
        if (!mounted.current || run !== generation.current) return;
        router.replace("/");
        router.refresh();
      }
    } catch (error) {
      if (
        requestController.signal.aborted ||
        !mounted.current ||
        run !== generation.current
      )
        return;
      setMessage(
        error instanceof Error
          ? error.message
          : "Admin sign-in could not continue.",
      );
    } finally {
      if (mounted.current && run === generation.current) setBusy(false);
    }
  }

  return (
    <form
      aria-busy={busy}
      aria-describedby={message ? "admin-login-error" : undefined}
      className="admin-login-form"
      onSubmit={(event) => {
        event.preventDefault();
        void submit(stage === "email" ? "start" : "complete");
      }}
    >
      <div className="admin-login-progress" aria-label="Sign-in progress">
        <span
          aria-current={stage === "email" ? "step" : undefined}
          className="is-active"
        >
          1
        </span>
        <i aria-hidden="true" className={stage === "code" ? "is-active" : ""} />
        <span
          aria-current={stage === "code" ? "step" : undefined}
          className={stage === "code" ? "is-active" : ""}
        >
          2
        </span>
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
        disabled={stage === "code" || busy}
        inputRef={emailField}
        fullWidth
        label="Admin email"
        onChange={(event) => setEmail(event.target.value)}
        required
        type="email"
        value={email}
      />
      <TextField
        autoComplete="current-password"
        disabled={stage === "code" || busy}
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
                    disabled={stage === "code" || busy}
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
            Contact support to recover access
          </a>
        </div>
      ) : null}
      {stage === "code" ? (
        <SegmentedOtpInput
          autoFocus
          disabled={busy}
          label="Six-digit code"
          onChange={setCode}
          required
          value={code}
        />
      ) : null}
      {message ? (
        <Alert id="admin-login-error" severity="error">
          {message}
        </Alert>
      ) : null}
      {busy ? (
        <AdminSkeleton variant="inline" label="Checking securely" />
      ) : null}
      <Button
        aria-describedby={message ? "admin-login-error" : undefined}
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
        {stage === "email" ? "Send sign-in code" : "Verify and enter"}
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
            setShowPassword(false);
            setMessage("");
            requestAnimationFrame(() => emailField.current?.focus());
          }}
          type="button"
        >
          Use another email
        </Button>
      ) : null}
    </form>
  );
}
