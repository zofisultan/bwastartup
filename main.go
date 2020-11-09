package main

import (
	"bwastartup/auth"
	"bwastartup/campaign"
	"bwastartup/handler"
	"bwastartup/helper"
	"bwastartup/user"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	dsn := "root:@tcp(127.0.0.1:3306)/db_startup?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal(err.Error())
	}

	userRepository := user.NewRepository(db)
	campaignRepository := campaign.NewRepository(db)

	// campaigns, _ := campaignRepository.FindByUserID(1)

	// fmt.Println(len(campaigns))
	// for _, campaign := range campaigns {
	// 	fmt.Println(campaign.Name)
	// 	if len(campaign.CampaignImages) > 0 {
	// 		fmt.Println(campaign.CampaignImages[0].FileName)
	// 	}
	// }

	userService := user.NewService(userRepository)
	campaignService := campaign.NewService(campaignRepository)
	authService := auth.NewService()

	// input := campaign.CreateCampaignInput{}
	// input.Name = "Pangalangan dana"
	// input.ShortDescription = "short"
	// input.Description = "looooo"
	// input.GoalAmount = 100000
	// input.Perks = "perks1, perks2, perks3"
	// inputUser, _ := userService.GetUserByID(1)
	// input.User = inputUser

	// _, err = campaignService.CreateCampaign(input)
	// if err != nil {
	// 	log.Fatal(err.Error())
	// }

	// campaigns, err := campaignService.FindCampaigns(0)
	// for _, campaign := range campaigns {
	// 	fmt.Println(campaign.Name)
	// 	if len(campaign.CampaignImages) > 0 {
	// 		fmt.Println(campaign.CampaignImages[0].FileName)
	// 	}
	// }

	// token, err := authService.ValidateToken("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo0fQ.O8fze3oVIP-7ATWjLpLljtSFlgUViFzu45dhrcyv8mc")

	// if err != nil {
	// 	fmt.Println("ERROR")
	// 	fmt.Println("ERROR")
	// }

	// if token.Valid {
	// 	fmt.Println("VALID")
	// 	fmt.Println("VALID")
	// } else {
	// 	fmt.Println("INVALID")
	// }

	//fmt.Println(authService.GenerateToken(1001))

	//userService.SaveAvatar(4, "images/1-profile.png")

	// input := user.LoginInput{
	// 	Email:    "asa@gmail.com",
	// 	Password: "password",
	// }

	// user, err := userService.Login(input)
	// if err != nil {
	// 	fmt.Println("Terjadi kesalahan")
	// 	fmt.Println(err.Error())
	// }

	// fmt.Println(user.Email)
	// fmt.Println(user.Name)

	// userByEmail, err := userRepository.FindByEmail("asa@gmail.com")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// if userByEmail.ID == 0 {
	// 	fmt.Println("User tidak di temukan")
	// } else {
	// 	fmt.Println(userByEmail.Name)
	// }

	userHandler := handler.NewUserHandler(userService, authService)
	campaignHanlder := handler.NewCampaignHandler(campaignService)

	router := gin.Default()
	router.Static("/images", "./images")

	api := router.Group("/api/v1")

	api.POST("/users", userHandler.RegisterUser)
	api.POST("/sessions", userHandler.Login)
	api.POST("/email_checkers", userHandler.CheckEmailAvailability)
	api.POST("/avatars", authMiddleware(authService, userService), userHandler.UploadAvatar)

	api.GET("/campaigns", campaignHanlder.GetCampaigns)
	api.GET("/campaigns/:id", campaignHanlder.GetCampaign)
	api.POST("/campaigns", authMiddleware(authService, userService), campaignHanlder.CreateCampaign)
	api.PUT("/campaigns/:id", authMiddleware(authService, userService), campaignHanlder.UpdateCampaign)
	router.Run()

	// Step
	// 1. Input User
	// 2. handler mapping input dari user -> struct input
	// 3. service : melakukan mapping dari struct input ke struct user
	// 4. repository

	// userInput := user.RegisterUserInput{}
	// userInput.Name = "tes simpan dari service"
	// userInput.Occupation = "terserah"
	// userInput.Email = "asa@gmail.com"
	// userInput.Password = "password"

	// userService.RegisterUser(userInput)

	// user := user.User{
	// 	Name: "tes simpan",
	// }

	// userRepository.Save(user)

	// fmt.Println("Connection to Database Success")

	// var users []user.User

	// db.Find(&users)
	// // length = len(users)
	// // fmt.Println(length)

	// for _, user := range users {
	// 	fmt.Println(user.Name)
	// 	fmt.Println(user.Occupation)
	// }

	// router := gin.Default()
	// router.GET("/handler", handler)
	// router.Run()
}

// func handler(c *gin.Context) {
// 	dsn := "root:@tcp(127.0.0.1:3306)/db_startup?charset=utf8mb4&parseTime=True&loc=Local"
// 	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// 	if err != nil {
// 		log.Fatal(err.Error())
// 	}

// 	var users []user.User
// 	db.Find(&users)

// 	c.JSON(http.StatusOK, users)
// }

// ===============kebutuhan untuk middleware
// 1. ambil nilai header authrization : bearer tokentoken
// 2. dari header authrization, kita ambil nilai token nya saja
// 3. kita validasi token
// 4. kita ambil user_id
// 5. ambil user dari db berdasarkan user_id lewat service
// 6. kita set context isi nya user

func authMiddleware(authService auth.Service, userService user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if !strings.Contains(authHeader, "Bearer") { // search kata "bearer"
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		// Bearer tokentoken
		tokenString := ""
		arrayToken := strings.Split(authHeader, " ") // arraytoken = {'bearer', 'tokentoken'}
		if len(arrayToken) == 2 {
			tokenString = arrayToken[1]
		}

		token, err := authService.ValidateToken(tokenString)
		if err != nil {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		claim, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		userID := int(claim["user_id"].(float64))

		user, err := userService.GetUserByID(userID)
		if err != nil {
			response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
			c.AbortWithStatusJSON(http.StatusUnauthorized, response)
			return
		}

		c.Set("currentUser", user)
	}
}

// func authMiddleware(c *gin.Context) {
// 	authHeader := c.GetHeader("Authorization")

// 	if !strings.Contains(authHeader, "Bearer") { // search kata "bearer"
// 		response := helper.APIResponse("Unauthorized", http.StatusUnauthorized, "error", nil)
// 		c.AbortWithStatusJSON(http.StatusUnauthorized, response)
// 		return
// 	}

// 	// Bearer tokentoken
// 	tokenString := ""
// 	arrayToken := strings.Split(authHeader, " ") // arraytoken = {'bearer', 'tokentoken'}
// 	if len(arrayToken) == 2 {
// 		tokenString = arrayToken[1]
// 	}
// }
