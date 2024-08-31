package routes

import (
	"github.com/Fesaa/Media-Provider/config"
	"github.com/Fesaa/Media-Provider/log"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func GetConfig(ctx *fiber.Ctx) error {
	return ctx.JSON(config.I())
}

func RemovePage(ctx *fiber.Ctx) error {
	indexS := ctx.Params("index")
	index, err := strconv.Atoi(indexS)
	if err != nil {
		log.Debug("Invalid index", "index", indexS, "error", err)
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid index"})
	}

	newCfg := config.I()
	if index < 0 || index >= len(newCfg.Pages) {
		log.Debug("Invalid index", "index", index)
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid index"})
	}
	newCfg.Pages = append(newCfg.Pages[:index], newCfg.Pages[index+1:]...)

	err = config.Save(newCfg)
	if err != nil {
		log.Error("Failed to save config", "error", err)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to save config"})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func AddPage(ctx *fiber.Ctx) error {
	var page config.Page
	err := ctx.BodyParser(&page)
	if err != nil {
		log.Error("Failed to parse request body", "error", err)
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	newCfg := config.I()
	newCfg.Pages = append(newCfg.Pages, page)

	err = config.Save(newCfg)
	if err != nil {
		log.Error("Failed to save config", "error", err)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to save config"})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func UpdatePage(ctx *fiber.Ctx) error {
	var page config.Page
	err := ctx.BodyParser(&page)
	if err != nil {
		log.Error("Failed to parse request body", "error", err)
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	indexS := ctx.Params("index")
	index, err := strconv.Atoi(indexS)
	if err != nil {
		log.Debug("Invalid index", "index", indexS, "error", err)
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid index"})
	}

	newCfg := config.I()
	if index < 0 || index >= len(newCfg.Pages) {
		log.Debug("Invalid index", "index", index)
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid index"})
	}

	newCfg.Pages[index] = page

	err = config.Save(newCfg)
	if err != nil {
		log.Error("Failed to save config", "error", err)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to save config"})
	}

	return ctx.SendStatus(fiber.StatusOK)
}

func MovePage(ctx *fiber.Ctx) error {
	oldIndexS := ctx.Params("oldIndex")
	oldIndex, err := strconv.Atoi(oldIndexS)
	if err != nil {
		log.Debug("Invalid old index", "index", oldIndexS, "error, err")
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid old index"})
	}

	newIndexS := ctx.Params("newIndex")
	newIndex, err := strconv.Atoi(newIndexS)
	if err != nil {
		log.Debug("Invalid new index", "index", newIndexS, "error, err")
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid new index"})
	}

	newCfg := config.I()
	if oldIndex < 0 || oldIndex >= len(newCfg.Pages) {
		log.Debug("Invalid old index", "index", oldIndex)
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid old index"})
	}

	if newIndex < 0 || newIndex >= len(newCfg.Pages) {
		log.Debug("Invalid new index", "index", newIndex)
		return ctx.Status(400).JSON(fiber.Map{"error": "Invalid new index"})
	}

	page := newCfg.Pages[oldIndex]
	newCfg.Pages = append(newCfg.Pages[:oldIndex], newCfg.Pages[oldIndex+1:]...)
	newCfg.Pages = append(newCfg.Pages[:newIndex], append([]config.Page{page}, newCfg.Pages[newIndex:]...)...)

	err = config.Save(newCfg)
	if err != nil {
		log.Error("Failed to save config", "error", err)
		return ctx.Status(500).JSON(fiber.Map{"error": "Failed to save config"})
	}

	return ctx.SendStatus(fiber.StatusOK)
}
