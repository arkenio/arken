swagger: '2.0'
info:
  title: Arken API
  version: "0.0.1"

# array of all schemes that your API supports
schemes:
  - http

# will be prefixed to all paths
basePath: /api/v1
produces:
  - application/json
paths:
  /services:
    get:
      summary: Gets the list of services
      parameters:
        - name: "status"
          in: "query"
          type: string
          enum: ["stopped","starting","started","error","passivated","stopping"]
      responses:
        200:
          description: The service
          schema:
            type: array
            items:
              $ref: '#/definitions/ServiceCluster'
    post:
      summary: Creates a service
      description: |
        Creates a service
      parameters:
        - in: "body"
          name: "body"
          description: "Service definition that has to be created"
          required: true
          schema:
            $ref: "#/definitions/ServiceForCreation"
      responses:
        200:
          description: The service
          schema:
            $ref: '#/definitions/ServiceCluster'


  /services/{serviceId}:
    get:
      summary: Shows a service's properties
      description: |
        Shows the service properties that expose it amongst other things its status.
      parameters:
        - name: serviceId
          in: path
          description: Id of the service
          required: true
          type: string

      responses:
        200:
          description: The service
          schema:
            $ref: '#/definitions/ServiceCluster'
        404:
          description: The service does not exist
          schema:
            $ref: '#/definitions/Error'

    put:
      summary: Start/Stop/Passivate the service
      description: |
        Start/Stop/Passivate the service
      parameters:
        - name: serviceId
          in: path
          description: Id of the service
          required: true
          type: string
        - name: action
          in: query
          description: Id of the service
          required: true
          type: string
          enum: ['start','stop','passivate']

      responses:
        200:
          description: The service
          schema:
            $ref: '#/definitions/ServiceCluster'
        404:
          description: The service does not exist
          schema:
            $ref: '#/definitions/Error'
    delete:
      summary: destroys a service
      parameters:
        - name: serviceId
          in: path
          description: Id of the service
          required: true
          type: string
      responses:
        200:
          description: The service
          schema:
            $ref: '#/definitions/ServiceCluster'
        404:
          description: The service does not exist
          schema:
            $ref: '#/definitions/Error'
definitions:
  ServiceCluster:
    type: object
    properties:
      name:
        type: string
        description: The name of the service cluster.
      instances:
        type: array
        description: the list of service instances.
        items:
          $ref: '#/definitions/Service'
  Service:
    type: object
    properties:
      index:
        type: string
      nodeKey:
        type: string
      location:
        $ref: '#/definitions/Location'
      domain:
        type: string
      name:
        type: string
      status:
        $ref: '#/definitions/Status'
      config:
        $ref: '#/definitions/ServiceConfig'

  ServiceForCreation:
    type: object
    properties:
      name:
        type: string
      domain:
        type: string
      config:
        $ref: '#/definitions/ServiceConfig'
    example:
      name: nxio-000001
      domain: test.devio
      config:
        rancherInfo:
          templateId: community:nuxeo:0


  Location:
    type: object
    properties:
      Host:
        type: string
      Port:
        type: integer

  Status:
    type: object
    properties:
      alive:
        type: string
      current:
        type: string
      expected:
        type: string

  ServiceConfig:
    type: object
    properties:
      robots:
        type: string
        description: |
          The content of the robots.txt that gogeta has to serve
      Environment:
        type: object
      RancherInfo:
        $ref: '#/definitions/RancherInfo'


  RancherInfo:
    type: object
    properties:
      templateId:
        type: string
        description: The Rancher template used for that service
        default: ""
      environmentId:
        type: string
        description: |
          The technical rancher Environment Id.
      environmentName:
        type: string
        description: The user readable Rancher environment name.
      location:
        $ref: '#/definitions/Location'
      healtState:
        type: string
        description: The Rancher given healthState of the service
      currentStatus:
        type: string
        description: the status computed from healthState
    example:
      templateId: community:nuxeo:0




  Error:
    type: object
    properties:
      code:
        type: integer
        format: int32
      message:
        type: string
      fields:
        type: string
