# Deploying the SSH edition to AWS

The goal: `ssh ssh.tatemccauley.com` drops anyone into the blog TUI. No accounts,
no shell — the [wish](https://github.com/charmbracelet/wish) server speaks the SSH
protocol, but the only thing behind a connection is a fresh Bubble Tea session.

## What gets deployed

One static binary. Posts and `home.md` are embedded at build time (`embed.go` at
the repo root), so **publishing a post to the SSH edition = rebuild + redeploy**.
There are no runtime file dependencies except the host key.

Build for Graviton (t4g instances are arm64):

```sh
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o blog-ssh ./tui
```

Run modes:

```sh
go run ./tui              # local TUI (what you had before)
go run ./tui -serve       # SSH server on 0.0.0.0:23234
ssh -p 23234 localhost    # connect to it
```

Flags: `-host`, `-port`, `-host-key` (key is auto-generated on first run).

## AWS architecture

```
visitor ──ssh:22──▶ [EIP / ssh.tatemccauley.com] ──▶ t4g.nano ──▶ blog-ssh (systemd, non-root)
you     ──SSM Session Manager (no public sshd at all)──▶ same instance
```

1. **EC2 t4g.nano**, Amazon Linux 2023 (arm64). ~$3/mo on-demand.
2. **IAM instance role** with the `AmazonSSMManagedInstanceCore` managed policy
   (plus S3 read on your deploy bucket, below). All admin access goes through
   SSM Session Manager, so the real sshd never needs to be reachable.
3. **Security group**: inbound TCP 22 from `0.0.0.0/0` and `::/0`. Nothing else.
4. **Disable the system sshd** — SSM replaces it and frees port 22 for wish:
   `sudo systemctl disable --now sshd`
5. **Elastic IP** + DNS **A record** for `ssh.tatemccauley.com`. The apex points
   at the website host, so the SSH edition lives on a subdomain; visitors run
   `ssh ssh.tatemccauley.com`.

Cost reality check: t4g.nano ≈ $3/mo + public IPv4 ≈ $3.65/mo → **≈ $7/mo**.
Lightsail's $5/mo bundle (IPv4 included) is the cheaper managed alternative if
you don't need raw EC2.

## Instance setup (one time, via SSM session)

```sh
sudo useradd --system --home-dir /var/lib/blog-ssh --shell /sbin/nologin blog
sudo mkdir -p /opt/blog /var/lib/blog-ssh
sudo chown blog:blog /var/lib/blog-ssh
```

`/etc/systemd/system/blog-ssh.service`:

```ini
[Unit]
Description=tatemccauley.com blog over SSH
After=network-online.target
Wants=network-online.target

[Service]
User=blog
Group=blog
ExecStart=/opt/blog/blog-ssh -serve -port 22 -host-key /var/lib/blog-ssh/host_ed25519
WorkingDirectory=/var/lib/blog-ssh
Restart=on-failure
RestartSec=2

# Bind port 22 without running as root
AmbientCapabilities=CAP_NET_BIND_SERVICE
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

# Belt and suspenders: the process can only write its own state dir
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/blog-ssh
PrivateTmp=true

[Install]
WantedBy=multi-user.target
```

Then: `sudo systemctl daemon-reload && sudo systemctl enable --now blog-ssh`

## Shipping the binary (deploys and post publishes)

No public sshd means no `scp`, so stage through S3 and apply with SSM:

```sh
# from this repo
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o blog-ssh ./tui
aws s3 cp blog-ssh s3://<deploy-bucket>/blog-ssh

aws ssm send-command \
  --instance-ids i-XXXXXXXX \
  --document-name AWS-RunShellScript \
  --parameters 'commands=[
    "aws s3 cp s3://<deploy-bucket>/blog-ssh /opt/blog/blog-ssh.new",
    "chmod 755 /opt/blog/blog-ssh.new",
    "mv /opt/blog/blog-ssh.new /opt/blog/blog-ssh",
    "systemctl restart blog-ssh"
  ]'
```

Worth automating as a GitHub Action on push to `main` once the manual flow feels
right — since posts are embedded, publishing requires it anyway.

If SSM feels heavy for a personal box, the fallback is real sshd on port 2222
with the security group restricted to *your* IP, and plain `scp -P 2222`.

## Host key

Auto-generated at `/var/lib/blog-ssh/host_ed25519` on first run. **Back it up**
(e.g. SSM Parameter Store, SecureString). If it changes, every returning visitor
gets an OpenSSH man-in-the-middle warning — the one thing you can't fix after
the fact.

## Security posture

- Anonymous by design: no passwords or accounts exist, so there is nothing to
  brute-force. Port-22 scanner noise is cosmetic; `journalctl -u blog-ssh` shows
  connects/disconnects if you're curious.
- Sessions idle out after 30 minutes and are capped at 100 concurrent
  (`tui/serve.go`) so a connection flood degrades into polite rejections.
- The service runs as a nologin system user with `CAP_NET_BIND_SERVICE` only,
  and systemd sandboxing limits writes to its own state directory.

## Smoke test

```sh
ssh ssh.tatemccauley.com                          # the real thing
ssh -o UserKnownHostsFile=/dev/null ssh.tatemccauley.com   # fresh-visitor view
```
