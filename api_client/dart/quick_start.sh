#!/bin/bash

# Flutter API客户端 - 快速测试脚本
# 使用方法: ./quick_start.sh

echo "🚀 Flutter API客户端 - 快速测试"
echo "=================================="

# 检查当前目录
if [ ! -f "pubspec.yaml" ]; then
    echo "❌ 错误: 请在包含 pubspec.yaml 的目录中运行此脚本"
    exit 1
fi

# 检查Flutter环境
if ! command -v flutter &> /dev/null; then
    echo "❌ 错误: Flutter未安装或不在PATH中"
    echo "请访问 https://flutter.dev/docs/get-started/install 安装Flutter"
    exit 1
fi

echo "📋 检查Flutter环境..."
flutter --version

# 安装依赖
echo ""
echo "📦 安装依赖包..."
flutter pub get

if [ $? -ne 0 ]; then
    echo "❌ 依赖安装失败"
    exit 1
fi

echo "✅ 依赖安装成功"

# 检查核心文件
echo ""
echo "🔍 检查核心文件..."

files=("api_client.dart" "models.dart" "test_config.dart")
for file in "${files[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file 存在"
    else
        echo "❌ $file 缺失"
        exit 1
    fi
done

# 运行示例代码
echo ""
echo "🧪 运行测试示例..."
echo "注意: 确保后端服务运行在 http://localhost:8089"
echo ""

# 检查是否有example.dart
if [ -f "example.dart" ]; then
    echo "执行示例代码..."
    dart example.dart
else
    echo "⚠️ example.dart 不存在，跳过示例测试"
fi

echo ""
echo "🎉 快速测试完成！"
echo ""
echo "📚 下一步操作:"
echo "1. 查看 README.md 了解详细用法"
echo "2. 修改 test_config.dart 中的服务器配置"
echo "3. 在你的Flutter项目中集成这些文件"
echo ""
echo "📝 文件说明:"
echo "- api_client.dart    : 核心API客户端"
echo "- models.dart        : 数据模型定义"
echo "- test_config.dart   : 测试配置"
echo "- example.dart       : 使用示例"
echo "- README.md          : 完整文档"
echo ""
echo "🔗 有用的命令:"
echo "flutter pub get      : 安装依赖"
echo "dart example.dart    : 运行示例"
echo "flutter run          : 运行Flutter应用"
echo ""
echo "✨ 祝你使用愉快！"
