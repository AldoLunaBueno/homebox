<template>
  <Dialog :open="isOpen" :dialog-id="DialogID.UploadInvoice" @update:open="setIsOpen">
    <DialogContent class="sm:max-w-md">
      <DialogHeader>
        <DialogTitle>Cargar Factura (PDF)</DialogTitle>
        <DialogDescription>
          Arrastra y suelta tu factura en formato PDF o haz clic para seleccionar el archivo.
        </DialogDescription>
      </DialogHeader>

      <div
        class="flex cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed p-8 text-center transition-colors"
        :class="isDragging ? 'border-primary bg-primary/10' : 'border-muted-foreground/25 hover:border-primary/50'"
        @dragover.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @drop.prevent="handleDrop"
        @click="triggerFileInput"
      >
        <MdiFileUpload class="mb-4 size-12 text-muted-foreground" />
        <p v-if="!selectedFile" class="text-sm text-muted-foreground">
          Arrastra el PDF aquí o <strong>haz clic</strong> para explorar
        </p>
        <p v-else class="text-sm font-medium text-primary">
          {{ selectedFile?.name }}
        </p>
        <input ref="fileInput" type="file" accept="application/pdf" class="hidden" @change="handleFileSelect" />
      </div>

      <DialogFooter class="mt-4 sm:justify-between">
        <Button variant="outline" @click="closeDialog">Cancelar</Button>
        <Button :disabled="!selectedFile || isUploading" @click="uploadInvoice">
          <MdiLoading v-if="isUploading" class="mr-2 animate-spin" />
          Procesar Factura
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>

<script setup lang="ts">
  import { ref, computed } from "vue";
  import { useDialog } from "~/components/ui/dialog-provider";
  import { DialogID } from "~/components/ui/dialog-provider/utils";
  import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
  } from "~/components/ui/dialog";
  import { Button } from "~/components/ui/button";
  import { toast } from "~/components/ui/sonner";
  import MdiFileUpload from "~icons/mdi/file-upload";
  import MdiLoading from "~icons/mdi/loading";

  const { activeDialog, closeDialog } = useDialog();
  const isOpen = computed(() => activeDialog.value === DialogID.UploadInvoice);

  const setIsOpen = (val: boolean) => {
    if (!val) {
      closeDialog();
      resetState();
    }
  };

  const fileInput = ref<HTMLInputElement | null>(null);
  const selectedFile = ref<File | null>(null);
  const isDragging = ref(false);
  const isUploading = ref(false);

  const resetState = () => {
    selectedFile.value = null;
    isDragging.value = false;
    isUploading.value = false;
  };

  const triggerFileInput = () => {
    if (fileInput.value) {
      fileInput.value.click();
    }
  };

  const handleDrop = (e: DragEvent) => {
    isDragging.value = false;
    const file = e.dataTransfer?.files?.[0];

    if (file && file.type === "application/pdf") {
      selectedFile.value = file;
    } else {
      toast.error("Por favor, sube únicamente un archivo PDF.");
    }
  };

  const handleFileSelect = (e: Event) => {
    const target = e.target as HTMLInputElement;
    const file = target.files?.[0];

    if (file) {
      selectedFile.value = file;
    }
  };

  const uploadInvoice = async () => {
    if (!selectedFile.value) return;

    isUploading.value = true;
    const formData = new FormData();
    formData.append("file", selectedFile.value);

    try {
      const response = await $fetch("/api/v1/invoices/process", {
        method: "POST",
        body: formData,
      });

      toast.success("Factura procesada con éxito");
      console.log("Datos de la factura:", response);

      closeDialog();
      resetState();
    } catch (error) {
      console.error(error);
      toast.error("Ocurrió un error al procesar el PDF.");
    } finally {
      isUploading.value = false;
    }
  };
</script>
