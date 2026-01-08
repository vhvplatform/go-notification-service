# Cấu trúc Repository Mới - Go Notification Service

## Tóm tắt

Repository đã được tổ chức lại theo cấu trúc monorepo với 3 thư mục chính:

```
go-notification-service/
├── server/          # Backend Golang microservice
├── client/          # Frontend ReactJS microservice (sẵn sàng để phát triển)
├── flutter/         # Mobile app Flutter (sẵn sàng để phát triển)
└── docs/           # Tài liệu chung của dự án
```

## Các thay đổi chính

### 1. Cấu trúc mới
- **server/**: Chứa toàn bộ code backend Golang (đã di chuyển từ root)
  - cmd/, internal/, go.mod, go.sum, Dockerfile, Makefile
  - Đã cập nhật Dockerfile để hoạt động với cấu trúc mới
  - Build và test đã được kiểm tra thành công

- **client/**: Thư mục dành cho ReactJS frontend
  - Có README.md mô tả cấu trúc dự kiến
  - Sẵn sàng để bắt đầu phát triển

- **flutter/**: Thư mục dành cho Flutter mobile app
  - Có README.md mô tả cấu trúc dự kiến
  - Sẵn sàng để bắt đầu phát triển

- **docs/**: Giữ nguyên tài liệu hiện có
  - DEPENDENCIES.md, PROVIDER_INTEGRATION.md, etc.
  - Các sơ đồ kiến trúc (diagrams/)

### 2. Cập nhật tài liệu
- README.md gốc đã được cập nhật để phản ánh cấu trúc mới
- Thêm README.md chi tiết cho từng thư mục con
- Hướng dẫn build và development đã được cập nhật

### 3. Kiểm tra
- ✅ Build thành công: `make build` trong thư mục server/
- ✅ Tất cả file đã được di chuyển đầy đủ
- ✅ Không có file bị mất hoặc bị xóa
- ✅ Dockerfile đã được cập nhật và hoạt động

## Lệnh Checkout

### Nếu đã có repository local:

```bash
# Cập nhật từ remote
git fetch origin

# Checkout nhánh mới
git checkout copilot/refactor-repository-structure

# Pull các thay đổi mới nhất
git pull origin copilot/refactor-repository-structure
```

### Nếu checkout mới (clone repository):

```bash
# Clone repository
git clone https://github.com/vhvplatform/go-notification-service.git

# Vào thư mục
cd go-notification-service

# Checkout nhánh mới
git checkout copilot/refactor-repository-structure
```

## Hướng dẫn phát triển

### Backend (Golang)

```bash
cd server

# Download dependencies
go mod download

# Build
make build

# Run tests
make test

# Run locally
make run
```

### Frontend (ReactJS) - Sẽ được thêm sau

```bash
cd client

# Sẽ được thêm khi bắt đầu phát triển
npm install
npm start
```

### Mobile App (Flutter) - Sẽ được thêm sau

```bash
cd flutter

# Sẽ được thêm khi bắt đầu phát triển
flutter pub get
flutter run
```

## Lưu ý quan trọng

1. **Không có dữ liệu bị mất**: Tất cả code và file cũ đã được di chuyển đầy đủ vào thư mục `server/`
2. **Build hoạt động bình thường**: Đã test và confirm build thành công
3. **Dockerfile đã được cập nhật**: Phù hợp với cấu trúc mới
4. **Tài liệu đã được cập nhật**: README và hướng dẫn phản ánh đúng cấu trúc mới

## Branch hiện tại

- **Branch name**: `copilot/refactor-repository-structure`
- **Latest commit**: Restructure repository: move backend to server/, add client/ and flutter/ directories
- **Status**: ✅ Ready for review and merge

## Tiếp theo

1. Review và merge branch này vào main
2. Bắt đầu phát triển client (ReactJS)
3. Bắt đầu phát triển flutter app
4. Cập nhật CI/CD pipeline nếu cần (để build từ thư mục server/)
