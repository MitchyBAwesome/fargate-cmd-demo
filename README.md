# AWS Fargate from the Command-Line

We all love a good command-line demo. Here is one I put together recently for demonstrating a few things, firstly Docker multi-stage builds, and secondly how a simple web service written in Go could be deployed to AWS Fargate using nothing but the command line.

## What's cool about that I hear you ask?
What's cool about it is that at no point during this demo am I deploying, configuring or going to have to manage ANY servers.

## Is it "serverless" ... is it containers? YES!
Let’s take a look: 

1. Create some environment variables
```
ACCOUNT_ID="000000000"

REGION="us-west-2"

REPO="go-http-server"
```
2. Start by creating an ECR (Elastic Container Registry) repository.
```
aws ecr create-repository --repository-name ${REGION}
```
3. Login to ecr
```
$(aws ecr get-login --no-include-email --region ${REGION})
```
4. Build docker image
```
docker build -t ${REPO} .
```
5. Tag docker image
```
docker tag ${REPO}:latest ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}:latest
```
6. Push docker image
```
docker push ${ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com/${REPO}:latest
```
7. Create Cluster
```
aws ecs create-cluster --cluster-name fargate-cluster --region ${REGION}
```
8. Get VPC Details
```
vpcid=$(aws ec2 describe-vpcs | jq '.Vpcs[] | select (.IsDefault == true) | .VpcId' | sed -e 's/^"//' -e 's/"$//')
```
9. Get subnet details
```
subnets=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=${vpcid}" | jq '.Subnets[].SubnetId') && echo $subnets
```
10. Create ELB security group
```
aws ec2 create-security-group \
--description "security group for ALB" \
--group-name "security-group-for-alb" \
--region ${REGION} | jq '.GroupId'
```
11. Configure ELB security group ingress rule to allow ELB to connect to tasks. 
```
aws ec2 authorize-security-group-ingress \
--group-id <GroupId_from_previous_command> \
--protocol tcp \
--port 80 \
--cidr 0.0.0.0/0
```
12. Create a load balancer and get the ELB ARN.
```
aws elbv2 create-load-balancer \ 
--name go-http-server \ 
--subnets subnet-111111 \
subnet-222222 \
subnet-333333 \ 
--security-groups <ALB_Security_GroupId> \
| jq '.LoadBalancers[].LoadBalancerArn'
```
13. Create a target group
```
aws elbv2 create-target-group \
--name fargate-targets \
--target-type ip \
--protocol HTTP \
--port 80 \
--vpc-id ${vpcid} | jq '.TargetGroups[].TargetGroupArn'
```
14. Create a listener
```
aws elbv2 create-listener \
--load-balancer-arn <load_balancer_arn> \
--protocol HTTP \
--port 80 \
--default-actions Type=forward,TargetGroupArn=<target_group_arn>
```
15. Register Task Definition
```
aws ecs register-task-definition \
--cli-input-json file://./go-http-server.json \
--region ${REGION} --query 'taskDefinition.taskDefinitionArn'
```
16. Create security group for the tasks
```
aws ec2 create-security-group \
--description "security group for fargate task" \
--group-name "security-group-for-fargate-task" \
--region ${REGION} | jq '.GroupId'
```
17. Configure security group ingress rule to allow ELB to connect to tasks.
```
aws ec2 authorize-security-group-ingress \
--group-id \
--protocol tcp \
--port 8000 \
--source-group <alb_security_group_id>
```
18. Create Service
```
aws ecs create-service --cluster fargate-cluster --service-name go-http-server \
--task-definition <task_definition> --desired-count 2 --launch-type "FARGATE" \
--network-configuration "awsvpcConfiguration={subnets=[ <command_separated_list_subnets> ],securityGroups=[<security_group>],assignPublicIp=ENABLED}" \
--load-balancers targetGroupArn=<target_group_arn>,containerName=<container_name>,containerPort=<container_port> \
--region ${REGION}
```
19. Grab the DNSName of the load balancer by querying the load balancer ARN.
```
url=$(aws elbv2 describe-load-balancers \
--load-balancer-arns <load_balancer_arn> \
| jq '.LoadBalancers[].DNSName' | sed -e 's/^"//' -e 's/"$//') && curl $url
```