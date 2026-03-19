package mail // nolint: revive

func emailVerificationTemplate() string {
	return `<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
</head>
<body style="margin:0; padding:0; background-color:#f4f4f7; font-family:Arial, Helvetica, sans-serif;">
  <table width="100%" cellpadding="0" cellspacing="0" style="background-color:#f4f4f7; padding:40px 0;">
    <tr>
      <td align="center">
        <table width="480" cellpadding="0" cellspacing="0" style="background-color:#ffffff; border-radius:8px; overflow:hidden; box-shadow:0 2px 8px rgba(0,0,0,0.08);">
          <!-- Header -->
          <tr>
            <td style="background-color:#1a1a2e; padding:24px; text-align:center;">
              <h1 style="margin:0; color:#ffffff; font-size:22px; font-weight:600;">ShortURL</h1>
            </td>
          </tr>
          <!-- Body -->
          <tr>
            <td style="padding:32px 28px;">
              <p style="margin:0 0 16px; color:#333333; font-size:15px; line-height:1.6;">Hi there,</p>
              <p style="margin:0 0 24px; color:#333333; font-size:15px; line-height:1.6;">
                We received a request to verify the email address associated with your account at <strong>{{ .ServiceURL }}</strong>. Enter the code below to complete your verification:
              </p>
              <!-- OTP Code -->
              <table width="100%" cellpadding="0" cellspacing="0">
                <tr>
                  <td align="center" style="padding:8px 0 24px;">
                    <div style="display:inline-block; background-color:#f0f0f5; border:2px dashed #1a1a2e; border-radius:8px; padding:16px 36px;">
                      <span style="font-size:32px; font-weight:700; letter-spacing:8px; color:#1a1a2e; font-family:'Courier New', monospace;">{{ .OTP }}</span>
                    </div>
                  </td>
                </tr>
              </table>
              <p style="margin:0 0 24px; color:#666666; font-size:13px; line-height:1.5; text-align:center;">
                This code expires in <strong>1 hour</strong>. Do not share it with anyone.
              </p>
              <hr style="border:none; border-top:1px solid #eeeeee; margin:24px 0;">
              <p style="margin:0; color:#999999; font-size:13px; line-height:1.5;">
                If you didn't create an account with ShortURL, you can safely ignore this email.
              </p>
            </td>
          </tr>
          <!-- Footer -->
          <tr>
            <td style="background-color:#f9f9fb; padding:20px 28px; text-align:center;">
              <p style="margin:0; color:#999999; font-size:12px;">&copy; {{ .Year }} ShortURL. All rights reserved.</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`
}
