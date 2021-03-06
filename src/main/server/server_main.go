package main

import (
	"context"
	"fmt"
	"github.com/DapperBlondie/go-grpc/src/messages/files"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"math"
	"net"
	"strconv"
	"time"
)

// GreetService this structure implement our RPC services
type GreetService struct{}

// SumService this structure implement our RPC services
type SumService struct{}

// GreetWithDeadline use for implementing deadline exceeded from client
func (gs *GreetService) GreetWithDeadline(ctx context.Context, request *files.GreetWithDeadlineRequest) (*files.GreetWithDeadlineResponse, error) {
	for i := 0; i < 3; i += 1 {
		if ctx.Err() == context.Canceled {
			log.Println("The client cancel the request !")
			return nil, status.Error(codes.Canceled, "The client canceled the request !")
		}
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("The deadline exceeded the request !")
			return nil, status.Error(codes.DeadlineExceeded, "The deadline exceeded the request !")
		}
		time.Sleep(time.Second * 1)
	}

	result := "Hello, " + request.GetGreet().GetFirstName() + ", " + request.GetGreet().GetLastName()
	res := &files.GreetWithDeadlineResponse{Result: result}
	return res, nil
}

// SquareRoot use for implementing an error that name is InvalidArgument for a Unary API
func (ss *SumService) SquareRoot(ctx context.Context, r *files.SquareRootRequest) (*files.SquareRootResponse, error) {
	fmt.Println("Get the number from request")
	number := r.GetNumber()
	if number < 0 {
		// This that error we will send to the client InvalidArgument
		return nil, status.Errorf(codes.InvalidArgument,
			fmt.Sprintf("%v is a negative number you sent", number))
	} else {
		resp := &files.SquareRootResponse{RootNumber: math.Sqrt(float64(number))}
		return resp, nil
	}
}

// EvenOrOdd for recognizing the request number is even or odd
func (ss *SumService) EvenOrOdd(stream files.SumService_EvenOrOddServer) error {
	resp := &files.NumResp{RespNum: "odd"}
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println(err.Error())
			return err
		} else if err != nil {
			log.Println(err.Error() + " occurred in EvenOdd API")
			return err
		}

		if (req.GetReqNum() % 2) == 1 {
			err := stream.Send(resp)
			if err != nil {
				log.Println(err.Error() + "Occurred in sending the response")
				return err
			}
		} else {
			resp.RespNum = "even"
			err := stream.Send(resp)
			if err != nil {
				log.Println(err.Error() + "Occurred in sending the response")
				return err
			}
		}
	}
}

// GreetEveryone use for greeting to the client requests
func (gs *GreetService) GreetEveryone(stream files.GreetService_GreetEveryoneServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return err
		} else if err != nil {
			log.Println("Error in receiving the req : " + err.Error())
			return err
		} else {
			firstName := req.GetGreeting().GetFirstName()
			result := "Hello " + firstName + " !"
			resp := &files.GreetEveryoneResponse{Result: result}
			sendErr := stream.Send(resp)
			if sendErr != nil {
				log.Println("Error during send response to the server : " + err.Error())
				return sendErr
			}
		}
	}
}

// AverageStreamingResult use for computing average of int32 client streaming
func (ss *SumService) AverageStreamingResult(stream files.SumService_AverageStreamingResultServer) error {
	var counter int32 = 0
	var sum int32 = 0

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Error in getting the request in Average Stream : " + err.Error())
			resp := &files.AverageResultResponse{Average: float32(sum) / float32(counter)}
			err := stream.SendAndClose(resp)
			return err
		} else if err != nil {
			log.Println("Error in Average Streaming : " + err.Error())
			return err
		} else {
			sum += req.GetNum()
			counter += 1
		}
	}
}

// LongGreet use for getting long greet then will send the number of greeting
func (gs *GreetService) LongGreet(stream files.GreetService_LongGreetServer) error {
	var counter int = 0
	var name string = ""

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Println("Error in LongGreet service : " + err.Error())
			resp := &files.LongGreetResponse{Result: name + ", Hello : " + strconv.Itoa(counter)}
			err := stream.SendAndClose(resp)
			return err
		} else if err != nil {
			//log.Println("Error in LongGreet service : " + err.Error())
			return err
		} else {
			name = req.GetGreeting().GetFirstName()
			counter += 1
		}
	}
}

// GetStreamingSumResult send the result of the stream of the repeated data
func (ss *SumService) GetStreamingSumResult(r *files.SumRequest, stream files.SumService_GetStreamingSumResultServer) error {
	lst := r.GetList()
	var number int32 = 0
	for _, value := range lst {
		number += value
		resp := &files.SumResponse{Result: number}
		err := stream.Send(resp)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 300)
	}

	return nil
}

// GetSumResult use for computing the sum of the repeated values from our request
func (ss *SumService) GetSumResult(ctx context.Context, r *files.SumRequest) (*files.SumResponse, error) {
	input := r.GetList()
	var result int32 = 0
	for _, num := range input {
		result += num
	}
	resp := &files.SumResponse{Result: result}

	return resp, nil
}

// GreetManyTimes use for stream many greet to our client
func (gs *GreetService) GreetManyTimes(r *files.GreetingManyTimeRequest, stream files.GreetService_GreetManyTimesServer) error {
	firstName := r.GetGreeting().GetFirstName()
	lastName := r.GetGreeting().GetLastName()

	result := firstName + " : " + lastName
	for i := 0; i < 10; i += 1 {
		result += " number " + strconv.Itoa(i) + "\n"
		resp := &files.GreetingManyTimesResponse{Result: result}
		err := stream.Send(resp)
		if err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 500)
	}
	return nil
}

//Greet our Greeting API just have on rpc service that we implemented that
func (gs *GreetService) Greet(ctx context.Context, r *files.GreetingRequest) (*files.GreetingResponse, error) {
	firstName := r.GetGreeting().FirstName
	lastName := r.GetGreeting().LastName

	result := fmt.Sprintf("Hello, %s %s\nI am Greeting API", firstName, lastName)

	resp := &files.GreetingResponse{Result: result}

	return resp, nil
}

func main() {
	err := runServer()
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	return
}

func runServer() error {
	log.Println("gRPC server is running ...")
	listener, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalln("Error in listening to rpc : " + err.Error())
		return err
	}

	srv := grpc.NewServer()
	files.RegisterGreetServiceServer(srv, &GreetService{})
	files.RegisterSumServiceServer(srv, &SumService{})
	reflection.Register(srv)

	log.Println("rpc Server listening on localhost:50051 ...")
	err = srv.Serve(listener)
	if err != nil {
		log.Fatalln("Error in serving our rpc server : " + err.Error())
		return err
	}

	return nil
}
