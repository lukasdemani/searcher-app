# Website Analyzer

A full-stack web application that crawls websites and analyzes their content, providing insights into HTML structure, link analysis, and login form detection.

## Features

- **URL Analysis**: Crawl websites and extract detailed information
- **HTML Version Detection**: Identify HTML version (HTML5, HTML 4.01, XHTML, etc.)
- **Heading Analysis**: Count heading tags by level (H1-H6)
- **Link Analysis**: Categorize links as internal/external and detect broken links
- **Login Form Detection**: Identify login forms on websites
- **Real-time Updates**: WebSocket-based status updates
- **Responsive Design**: Mobile and desktop-friendly interface
- **Bulk Operations**: Analyze or delete multiple URLs at once

## Tech Stack

### Backend
- **Go** (Golang) with Gin framework
- **MySQL** database
- **WebSocket** for real-time updates
- **Worker Pool** for concurrent processing
- **Docker** for containerization

### Frontend
- **React 19** with TypeScript
- **Tailwind CSS** for styling
- **Recharts** for data visualization
- **React Router** for navigation
- **i18next** for internationalization

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- Docker and Docker Compose
- MySQL 8.0 (or Docker)

## Quick Start

### 1. Clone the repository
```bash
git clone <repository-url>
cd searcher-app
```

### 2. Environment Setup
```bash
# Make the development script executable
chmod +x dev.sh

# Setup development environment
./dev.sh setup
```

### 3. Start the application
```bash
# Start all services (database, backend, frontend)
./dev.sh start
```

### 4. Access the application
- Frontend: http://localhost:5173
- Backend API: http://localhost:8080
- API Health Check: http://localhost:8080/health

## Manual Setup

### Database Setup
```bash
# Start MySQL with Docker
docker-compose up -d mysql

# Or connect to your existing MySQL instance
# Update environment variables in docker-compose.yml
```

### Backend Setup
```bash
cd backend

# Install dependencies
go mod download

# Set environment variables
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=analyzer_user
export DB_PASSWORD=analyzer_pass
export DB_NAME=website_analyzer
export API_KEY=dev-api-key-2024

# Run migrations (if available)
# mysql -u analyzer_user -p analyzer_pass website_analyzer < migrations/001_initial_schema.sql

# Start the backend server
go run cmd/main.go
```

### Frontend Setup
```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

## API Documentation

### Authentication
All API endpoints require an API key in the `X-API-Key` header:
```
X-API-Key: dev-api-key-2024
```

### Endpoints

#### URLs
- `GET /api/urls` - List URLs with pagination and filtering
- `POST /api/urls` - Add a new URL for analysis
- `GET /api/urls/:id` - Get URL details
- `PUT /api/urls/:id/analyze` - Start URL analysis
- `DELETE /api/urls/:id` - Delete a URL
- `GET /api/urls/:id/broken-links` - Get broken links for a URL

#### Bulk Operations
- `POST /api/urls/bulk-analyze` - Analyze multiple URLs
- `POST /api/urls/bulk-delete` - Delete multiple URLs

#### WebSocket
- `GET /ws` - WebSocket connection for real-time updates

### Example Usage

#### Add a URL
```bash
curl -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -H "X-API-Key: dev-api-key-2024" \
  -d '{"url": "https://example.com"}'
```

#### Get URLs
```bash
curl http://localhost:8080/api/urls?page=1&limit=10 \
  -H "X-API-Key: dev-api-key-2024"
```

## Data Collection

The analyzer collects the following data for each URL:

### Basic Information
- **Title**: Page title from `<title>` tag
- **HTML Version**: Detected from DOCTYPE declaration
- **Status**: Processing status (queued, processing, completed, error)

### Heading Analysis
- **H1-H6 Counts**: Number of each heading level

### Link Analysis
- **Internal Links**: Links to the same domain
- **External Links**: Links to different domains
- **Broken Links**: Links returning 4xx or 5xx status codes

### Form Detection
- **Login Forms**: Forms with username/password fields

## Development

### Project Structure
```
searcher-app/
├── backend/
│   ├── cmd/main.go              # Application entry point
│   ├── internal/
│   │   ├── database/            # Database connection
│   │   ├── handlers/            # HTTP handlers
│   │   ├── middleware/          # Authentication, etc.
│   │   ├── models/              # Data models
│   │   ├── repository/          # Data access layer
│   │   ├── services/            # Business logic
│   │   └── worker/              # Background job processing
│   └── migrations/              # Database migrations
├── frontend/
│   ├── src/
│   │   ├── components/          # React components
│   │   ├── hooks/               # Custom hooks
│   │   ├── pages/               # Page components
│   │   ├── services/            # API services
│   │   └── types/               # TypeScript types
│   └── public/                  # Static assets
└── docker-compose.yml           # Docker configuration
```

### Running Tests
```bash
# Frontend tests
cd frontend
npm test

# Backend tests (when available)
cd backend
go test ./...
```

### Building for Production
```bash
# Build backend
cd backend
go build -o analyzer cmd/main.go

# Build frontend
cd frontend
npm run build
```

## Docker Deployment

### Using Docker Compose
```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Environment Variables

#### Backend
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 3306)
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `API_KEY`: API key for authentication
- `PORT`: Server port (default: 8080)

#### Frontend
- `REACT_APP_API_BASE_URL`: Backend API URL
- `REACT_APP_WS_URL`: WebSocket URL

## Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Ensure MySQL is running
   - Check database credentials
   - Verify network connectivity

2. **API Key Authentication Failed**
   - Verify the API key is correct
   - Ensure the header is properly set

3. **WebSocket Connection Issues**
   - Check if the WebSocket URL is correct
   - Verify firewall settings

4. **Crawling Timeout**
   - Some websites may be slow to respond
   - Check network connectivity
   - Increase timeout values if needed

### Development Tools

```bash
# Health check
curl http://localhost:8080/health

# Database query
mysql -u analyzer_user -p analyzer_pass website_analyzer -e "SELECT * FROM urls;"

# View logs
docker-compose logs backend
docker-compose logs frontend
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

## 📊 Charts and Visualizations

### Interactive Charts in Details View

The application features comprehensive data visualizations in the URL details page:

#### **Links Distribution Chart**
- **Bar Chart**: Clean visualization of internal vs external links
- **Donut Chart**: Circular representation with percentages
- **Responsive**: Adapts to different screen sizes
- **Interactive**: Hover tooltips with detailed information
- **Multilingual**: Full support for EN/DE translations

#### **Headings Distribution Chart**
- **Bar Chart**: Distribution of H1-H6 heading tags
- **Dynamic Data**: Real-time data from website analysis
- **Professional Styling**: Consistent color scheme and typography
- **Accessible**: Proper contrast and descriptive labels

#### **Technical Implementation**
- **Recharts Library**: Professional React chart library
- **TypeScript**: Full type safety for chart components
- **Responsive Containers**: Automatic sizing and scaling
- **Custom Tooltips**: Enhanced user experience
- **Internationalization**: Complete i18n support
