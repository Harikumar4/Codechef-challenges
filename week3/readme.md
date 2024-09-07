## Endpoints

### 1. **Generate HTML**

- **Endpoint**: /generate-html
- **Method**: POST
- **Description**: Generates HTML content based on a provided template and content.

#### **Request**

- **Content-Type**: application/json
- **Request Body**:
{
    "template": "<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>Page Title</title></head><body><h1>Welcome</h1><div>{{content}}</div></body></html>",
    "content": "<p>Here is some content to include in the HTML.</p>"
}

### 2. **Modify HTML**

- **Endpoint**: `/modify-html`
- **Method**: `POST`
- **Description**: Modifies an existing HTML content based on provided instructions.

#### **Request**

- **Content-Type**: `application/json`
- **Request Body**:
  {
    "html": "<!DOCTYPE html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>Page Title</title></head><body><h1>Welcome</h1><div><p>Current content</p></div></body></html>",
    "instructions": "Add a new section about upcoming events after the existing content."
}
