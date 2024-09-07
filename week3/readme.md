## Endpoints

### 1. **Generate HTML**

- **Endpoint**: /generate-html
- **Method**: POST
- **Description**: Generates HTML content based on a provided template and content.

#### **Request**

- **Content-Type**: application/json
- **Request Body**:
{
    "template": "any html basic template>",
    "content": "you want to generate about"
}

### 2. **Modify HTML**

- **Endpoint**: `/modify-html`
- **Method**: `POST`
- **Description**: Modifies an existing HTML content based on provided instructions.

#### **Request**

- **Content-Type**: `application/json`
- **Request Body**:
  {
    "html": "existing html template",
    "instructions": "any modification you would like in it"
}
