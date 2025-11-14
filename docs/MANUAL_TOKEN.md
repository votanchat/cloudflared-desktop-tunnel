# Sử dụng Manual Token

## Tổng quan

App hỗ trợ 2 cách lấy token để start tunnel:

1. **Tự động từ Backend** (mặc định) - App sẽ gọi API backend để lấy token
2. **Manual Token** - Bạn tự paste token vào UI

## Khi nào dùng Manual Token?

✅ **Nên dùng khi:**
- Testing/development mà chưa có backend
- Muốn kiểm soát token cụ thể
- Backend tạm thời không khả dụng
- Debug tunnel connection issues

❌ **Không nên dùng trong production:**
- Không secure để user paste token trực tiếp
- Khó quản lý token rotation
- Không có centralized control

## Cách lấy Cloudflare Tunnel Token

### Option 1: Cloudflare Dashboard

1. Đăng nhập [Cloudflare Zero Trust Dashboard](https://one.dash.cloudflare.com/)
2. Vào **Networks** → **Tunnels**
3. Click vào tunnel của bạn
4. Tab **Configure**
5. Scroll xuống phần **Connector** → Click **View token**
6. Copy token

### Option 2: CLI (Khuyến nghị cho dev)

```bash
# 1. Login Cloudflare
cloudflared tunnel login

# 2. Tạo tunnel mới (nếu chưa có)
cloudflared tunnel create my-test-tunnel

# Output sẽ hiển thị tunnel ID
# Tunnel credentials written to /Users/you/.cloudflared/<UUID>.json

# 3. Lấy token
cloudflared tunnel token <TUNNEL_ID>

# hoặc dùng tunnel name
cloudflared tunnel token my-test-tunnel

# Copy token từ output
```

### Token Format

Token có dạng:
```
eyJhIjoiMTIzNDU2Nzg5MGFiY2RlZiIsInQiOiJhYmNkZWYxMi0zNDU2LTc4OTAtYWJjZC1lZjEyMzQ1Njc4OTAiLCJzIjoiWldGaFpHVm1NVEl6TkRVMk56ZzVNR0ZpWTJSbFpnPT0ifQ==
```

Đây là base64-encoded JSON chứa:
- Account ID
- Tunnel ID  
- Secret

## Cách sử dụng trong App

### 1. Mở Tab "Tunnel"

### 2. Click nút "✏️ Manual Token"

Section "🔑 Token Configuration" sẽ mở ra.

### 3. Paste token vào textarea

```
eyJhIjoiMTIzNDU2Nzg5MGFiY2RlZiIsInQiOiJhYmNkZWYxMi0zNDU2LTc4OTAtYWJjZC1lZjEyMzQ1Njc4OTAiLCJzIjoiWldGaFpHVm1NVEl6TkRVMk56ZzVNR0ZpWTJSbFpnPT0ifQ==
```

### 4. Click "▶️ Start Tunnel"

App sẽ:
- ✅ Sử dụng token bạn vừa paste
- ✅ **KHÔNG** gọi backend API
- ✅ Start cloudflared với token đó
- ✅ Tự động xóa token khỏi input sau khi start (security)

### 5. Để dùng backend token

Chỉ cần **không nhập gì** vào Manual Token field, hoặc click "❌ Hide" để ẩn input.

Khi start tunnel:
- ✅ App sẽ gọi `GET /api/token` từ backend
- ✅ Dùng token từ backend response

## Flow Diagram

```
User clicks "Start Tunnel"
         |
         v
Manual Token field có giá trị?
         |
    +----+----+
    |         |
   YES       NO
    |         |
    v         v
Dùng      Gọi Backend
Manual    GET /api/token
Token          |
    |          v
    |     Dùng Backend
    |        Token
    |          |
    +----+-----+
         |
         v
   Start cloudflared
   với token
```

## Security Notes

### ✅ Good Practices

1. **Token chỉ dùng cho development/testing**
2. **Không commit token vào Git**
3. **Token tự động xóa khỏi UI sau khi start**
4. **Token không được lưu vào config file**
5. **Token không được log ra console**

### ❌ Không nên

1. ❌ Share token qua chat/email
2. ❌ Hardcode token trong code
3. ❌ Dùng production token cho testing
4. ❌ Để token trong clipboard lâu

## Troubleshooting

### Token không valid

**Error:** `failed to start tunnel: Invalid tunnel token`

**Solutions:**
- Kiểm tra token có đầy đủ không (thường rất dài)
- Đảm bảo không có space/newline thừa
- Lấy token mới từ Cloudflare
- Kiểm tra tunnel vẫn còn tồn tại trong Cloudflare

### Token expired

**Error:** `authentication failed`

Tokens có thể expire nếu:
- Tunnel bị xóa
- Tunnel credentials bị revoke
- Account permissions thay đổi

**Solution:** Lấy token mới từ Cloudflare Dashboard hoặc CLI.

### Backend không khả dụng

**Error:** `failed to fetch token from backend: connection refused`

Đây chính là lúc Manual Token hữu ích:
1. Click "✏️ Manual Token"
2. Paste token từ Cloudflare
3. Start tunnel

## Example: Testing với Manual Token

```bash
# 1. Lấy token
cloudflared tunnel token my-dev-tunnel

# Output:
eyJhIjoiMTIzNDU2Nzg5MGFiY2RlZiIsInQiOiJhYmNkZWYxMi0zNDU2LTc4OTAtYWJjZC1lZjEyMzQ1Njc4OTAiLCJzIjoiWldGaFpHVm1NVEl6TkRVMk56ZzVNR0ZpWTJSbFpnPT0ifQ==

# 2. Copy token

# 3. Mở app → Tab Tunnel → Manual Token

# 4. Paste token

# 5. Start tunnel
```

## Kiểm tra tunnel đang chạy

```bash
# Kiểm tra process
ps aux | grep cloudflared

# Kiểm tra logs trong app UI
# hoặc
tail -f /tmp/cloudflared-*.log
```

## Production Deployment

Trong production:

1. **Disable manual token input** (hoặc hide UI)
2. **Luôn dùng backend API** để fetch token
3. **Implement token rotation** ở backend
4. **Monitor token expiry** và auto-refresh
5. **Log token usage** cho audit trail

## Code Example: Backend API

Nếu bạn muốn build backend để issue tokens:

```javascript
// Node.js + Express example
app.get('/api/token', async (req, res) => {
  // Authenticate user
  const userId = req.user.id;
  
  // Get tunnel ID for this user
  const tunnelId = await getTunnelForUser(userId);
  
  // Generate token from Cloudflare API
  const token = await cloudflare.tunnels.getToken(tunnelId);
  
  // Return token with expiry
  res.json({
    token: token,
    expiresAt: new Date(Date.now() + 24 * 60 * 60 * 1000)
  });
});
```

## FAQ

**Q: Token có expire không?**
A: Tunnel tokens thường không expire trừ khi tunnel bị xóa hoặc credentials bị revoke.

**Q: Có thể dùng chung 1 token cho nhiều máy?**
A: Có, nhưng không khuyến nghị. Nên tạo tunnel riêng cho mỗi client.

**Q: Token lưu ở đâu sau khi start?**
A: Token **KHÔNG** được lưu. Nó chỉ dùng để start process cloudflared, sau đó bị xóa khỏi memory.

**Q: Manual token có được gửi lên backend không?**
A: **KHÔNG**. Khi dùng manual token, app không gọi backend API.

## Related Docs

- [Backend API Specification](./BACKEND_API.md)
- [Architecture Documentation](./ARCHITECTURE.md)
- [Setup Guide](../SETUP.md)
- [Cloudflare Tunnel Docs](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/)
