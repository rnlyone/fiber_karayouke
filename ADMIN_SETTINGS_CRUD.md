# Admin Settings CRUD Implementation

Complete CRUD (Create, Read, Update, Delete) functionality for system configuration settings.

## Backend Changes

### Endpoints (app/routes/web.go)
- **GET** `/api/admin/configs` - Read all configurations
- **POST** `/api/admin/configs` - Create new configuration
- **PUT** `/api/admin/configs/:key` - Update existing configuration
- **DELETE** `/api/admin/configs/:key` - Delete configuration

### Controller Methods (app/controllers/admin_controller.go)

#### 1. GetConfigs() - READ
- Returns all system configurations as a key-value map
- Includes default values for critical configs if not in database

#### 2. CreateConfig() - CREATE
- Creates a new configuration entry
- Validates key and value are not empty
- Returns 409 error if config with same key already exists
- Returns 201 status on success

#### 3. UpdateConfig() - UPDATE
- Updates existing configuration value
- Uses URL parameter `:key` to identify config
- Returns 404 if config not found
- Validates value is not empty

#### 4. DeleteConfig() - DELETE
- Deletes configuration by key
- **Protected**: Prevents deletion of critical system configs:
  - `room_max_duration`
  - `room_creation_cost`
  - `default_credits`
- Returns 400 error if trying to delete critical config
- Returns 404 if config not found

## Frontend Changes

### AdminSettings Component Features

#### 1. **Read (Display)**
- Shows all configurations in a list format
- Displays critical configs with warning badge
- Groups by predefined labels or shows raw key for custom configs

#### 2. **Create (Add New)**
- Toggle-able "Add Config" form
- Input fields for config key and value
- Validates both fields are filled
- Shows success/error messages
- Automatically refreshes list after creation

#### 3. **Update (Edit)**
- Inline editing for all config values
- Save button appears when value is changed
- Shows "Saving..." state during update
- Success/error feedback
- Auto-refresh after successful update

#### 4. **Delete (Remove)**
- Delete button for non-critical configs
- Confirmation dialog before deletion
- Critical configs cannot be deleted (no delete button)
- Success/error feedback

### UI/UX Improvements
- Success and error message banners with dismiss button
- Loading states on all actions
- Disabled state for buttons during operations
- Clear visual distinction for critical configs
- Responsive form layout

## Security Features
- All endpoints protected by AdminMiddleware
- Only users with admin email can access
- Critical configs cannot be deleted
- Input validation on both frontend and backend

## Usage Example

### Creating a Custom Config
1. Click "+ Add Config" button
2. Enter key: `max_playlist_size`
3. Enter value: `50`
4. Click "Create Config"

### Updating a Config
1. Change the value in the input field
2. Click "Save" button
3. Wait for success message

### Deleting a Config
1. Click "Delete" button (only visible for non-critical configs)
2. Confirm deletion in dialog
3. Config is removed from list

## Critical Configurations
These configs cannot be deleted but can be updated:
- **room_max_duration**: Room expiration time in minutes
- **room_creation_cost**: Credits required to create a room
- **default_credits**: Initial credits for new users
